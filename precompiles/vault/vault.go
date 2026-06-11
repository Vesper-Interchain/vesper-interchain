// Package vault implements a stateful EVM precompile that exposes the Vesper
// collateral vault and rewards functionality to Solidity contracts.
//
// The precompile lives at the well-known address 0x...0900 on the Vesper chain.
// It bridges the EVM execution environment to the Cosmos-side x/collateral and
// x/rewards keepers using the cosmos/evm stateful precompile pattern.
//
// Address equivalence: an EVM caller's address (20 bytes) maps 1:1 to a Cosmos
// AccAddress using the same raw bytes, so both representations control the same account.
// This means a user who deposits via the precompile and a user who deposits via the
// Cosmos MsgDepositCollateral both operate on the same vault state.
package vault

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	collateraltypes "github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// VaultPrecompileAddress is the canonical EVM address of the vault precompile.
// Registered as a static precompile in app/evm.go so all contracts on the chain
// can call it at this well-known address.
const VaultPrecompileAddress = "0x0000000000000000000000000000000000000900"

// abiJSON is the compiled ABI embedded from abi.json at build time.
// The file defines all function signatures and event definitions that Solidity
// callers use to encode calls and decode return values.
//
//go:embed abi.json
var abiJSON []byte

// ABI method name constants. These must match the function names in abi.json exactly.
const (
	DepositCollateralMethod  = "depositCollateral"
	WithdrawCollateralMethod = "withdrawCollateral"
	MintStablecoinMethod     = "mintStablecoin"
	RepayDebtMethod          = "repayDebt"
	LiquidatePositionMethod  = "liquidatePosition"
	ClaimRewardsMethod       = "claimRewards"
	GetPositionMethod        = "getPosition"
	GetPendingRewardsMethod  = "getPendingRewards"
)

// CollateralKeeper defines the collateral keeper interface required by this precompile.
// Using a local interface instead of importing the concrete keeper avoids a circular
// dependency between the precompile and the keeper packages.
type CollateralKeeper interface {
	DepositCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error
	WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error
	MintStablecoin(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error
	RepayDebt(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error
	LiquidatePosition(ctx sdk.Context, owner sdk.AccAddress, liquidator sdk.AccAddress) error
	GetPosition(ctx sdk.Context, owner string) (collateraltypes.Position, error)
	GetCollateralRatio(ctx sdk.Context, position collateraltypes.Position) (math.LegacyDec, error)
}

// RewardsKeeper defines the rewards keeper interface required by this precompile.
type RewardsKeeper interface {
	ClaimRewards(ctx sdk.Context, owner sdk.AccAddress) (math.Int, error)
	GetPendingRewards(ctx sdk.Context, owner sdk.AccAddress) math.Int
}

// Compile-time check that Precompile satisfies the vm.PrecompiledContract interface.
var _ vm.PrecompiledContract = &Precompile{}

// Precompile is the vault stateful precompile. It embeds cmn.Precompile which
// provides ABI lookup, gas accounting, balance change tracking, and the RunSetup
// helper that decodes calldata and extracts the SDK context from the EVM.
type Precompile struct {
	cmn.Precompile
	collateralKeeper CollateralKeeper
	rewardsKeeper    RewardsKeeper
}

// NewPrecompile parses the embedded ABI and constructs a new Precompile instance.
// bk (bank keeper) is passed to the embedded cmn.Precompile so it can track
// balance changes made by Cosmos-side operations that affect EVM balance views.
func NewPrecompile(ck CollateralKeeper, rk RewardsKeeper, bk cmn.BankKeeper) (*Precompile, error) {
	parsedABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault precompile ABI: %w", err)
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  parsedABI,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
		},
		collateralKeeper: ck,
		rewardsKeeper:    rk,
	}
	p.SetAddress(common.HexToAddress(VaultPrecompileAddress))
	p.SetBalanceHandler(bk)
	return p, nil
}

// RequiredGas returns the intrinsic gas cost for a precompile call based on
// whether the method is a transaction (state-modifying) or a query (read-only).
// Returns 0 for malformed input shorter than a 4-byte method selector.
func (p Precompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}
	method, err := p.MethodById(input[:4])
	if err != nil {
		return 0
	}
	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

// Run is the entry point called by the EVM for every precompile invocation.
// It delegates to run() and converts any error into an EVM revert so the
// calling contract receives a clean revert rather than a panic.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	bz, err = p.run(evm, contract, readOnly)
	if err != nil {
		return cmn.ReturnRevertError(evm, err)
	}
	return bz, nil
}

// run decodes the calldata, dispatches to the appropriate handler, and handles
// gas accounting and balance change propagation.
func (p Precompile) run(evm *vm.EVM, contract *vm.Contract, readOnly bool) ([]byte, error) {
	// RunSetup decodes the ABI calldata, extracts the SDK context from the EVM's
	// state database, and validates read-only constraints for query methods.
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// Snapshot the EVM balance state before any Cosmos-side changes so that
	// the balance handler can apply the net delta to the EVM world-state afterwards.
	p.GetBalanceHandler().BeforeBalanceChange(ctx)
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	var out []byte
	switch method.Name {
	case DepositCollateralMethod:
		out, err = p.depositCollateral(ctx, contract, stateDB, method, args)
	case WithdrawCollateralMethod:
		out, err = p.withdrawCollateral(ctx, contract, stateDB, method, args)
	case MintStablecoinMethod:
		out, err = p.mintStablecoin(ctx, contract, stateDB, method, args)
	case RepayDebtMethod:
		out, err = p.repayDebt(ctx, contract, stateDB, method, args)
	case LiquidatePositionMethod:
		out, err = p.liquidatePosition(ctx, contract, stateDB, method, args)
	case ClaimRewardsMethod:
		out, err = p.claimRewards(ctx, contract, stateDB, method, args)
	case GetPositionMethod:
		out, err = p.getPosition(ctx, method, args)
	case GetPendingRewardsMethod:
		out, err = p.getPendingRewards(ctx, method, args)
	default:
		return nil, vm.ErrExecutionReverted
	}
	if err != nil {
		return nil, err
	}

	// Deduct the gas consumed by Cosmos-side operations from the contract's gas budget.
	cost := ctx.GasMeter().GasConsumed() - initialGas
	if !contract.UseGas(cost, nil, tracing.GasChangeCallPrecompiledContract) {
		return nil, vm.ErrOutOfGas
	}

	// Propagate any Cosmos-side balance changes (mints, burns, transfers) back
	// into the EVM state database so the EVM world-state stays consistent.
	if err = p.GetBalanceHandler().AfterBalanceChange(ctx, stateDB); err != nil {
		return nil, err
	}

	return out, nil
}

// IsTransaction classifies a method as state-modifying (true) or read-only (false).
// Read-only methods are permitted in staticcall contexts and do not consume write gas.
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case DepositCollateralMethod, WithdrawCollateralMethod,
		MintStablecoinMethod, RepayDebtMethod,
		LiquidatePositionMethod, ClaimRewardsMethod:
		return true
	}
	return false
}

// callerToCosmosAddr converts an Ethereum address to a Cosmos AccAddress.
// Both types represent the same 20-byte address; only the encoding differs.
func callerToCosmosAddr(addr common.Address) sdk.AccAddress {
	return sdk.AccAddress(addr.Bytes())
}
