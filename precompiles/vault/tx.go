package vault

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/evm/x/vm/statedb"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// depositCollateral handles the depositCollateral(uint256 amount) transaction.
// The caller's EVM address is mapped to a Cosmos AccAddress (same 20 bytes) and
// the uatom amount is forwarded to the collateral keeper. Returns the caller's
// updated total collateral balance after the deposit.
func (p Precompile) depositCollateral(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, errInvalidArgs(method.Name, 1, len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok || amount == nil {
		return nil, errInvalidArgType(method.Name, "amount")
	}

	owner := callerToCosmosAddr(contract.Caller())
	cosmosAmount := math.NewIntFromBigInt(amount)

	if err := p.collateralKeeper.DepositCollateral(ctx, owner, cosmosAmount); err != nil {
		return nil, err
	}

	// Read back the updated position to return the new collateral total to the caller.
	pos, err := p.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		return nil, err
	}
	collateralAmount, _ := math.NewIntFromString(pos.CollateralAmount)
	return method.Outputs.Pack(collateralAmount.BigInt())
}

// withdrawCollateral handles the withdrawCollateral(uint256 amount) transaction.
// Returns the updated collateral ratio scaled by 1e18 (matching Solidity's
// uint256 representation for decimal values). Returns 0 if the position was closed.
func (p Precompile) withdrawCollateral(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, errInvalidArgs(method.Name, 1, len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok || amount == nil {
		return nil, errInvalidArgType(method.Name, "amount")
	}

	owner := callerToCosmosAddr(contract.Caller())
	if err := p.collateralKeeper.WithdrawCollateral(ctx, owner, math.NewIntFromBigInt(amount)); err != nil {
		return nil, err
	}

	// If the position was fully closed after withdrawal, GetPosition returns an error.
	// Return a ratio of 0 in that case to indicate no remaining position.
	pos, err := p.collateralKeeper.GetPosition(ctx, owner.String())
	var ratioBig *big.Int
	if err != nil {
		ratioBig = big.NewInt(0)
	} else {
		ratio, _ := p.collateralKeeper.GetCollateralRatio(ctx, pos)
		// Scale to 1e18 to match Solidity's convention for representing decimals as integers.
		ratioBig = ratio.MulInt64(1e18).TruncateInt().BigInt()
	}
	return method.Outputs.Pack(ratioBig)
}

// mintStablecoin handles the mintStablecoin(uint256 amount) transaction.
// Returns the minted amount and the updated collateral ratio (×1e18).
func (p Precompile) mintStablecoin(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, errInvalidArgs(method.Name, 1, len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok || amount == nil {
		return nil, errInvalidArgType(method.Name, "amount")
	}

	owner := callerToCosmosAddr(contract.Caller())
	cosmosAmount := math.NewIntFromBigInt(amount)

	if err := p.collateralKeeper.MintStablecoin(ctx, owner, cosmosAmount); err != nil {
		return nil, err
	}

	pos, err := p.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		return nil, err
	}
	ratio, _ := p.collateralKeeper.GetCollateralRatio(ctx, pos)
	ratioBig := ratio.MulInt64(1e18).TruncateInt().BigInt()

	return method.Outputs.Pack(amount, ratioBig)
}

// repayDebt handles the repayDebt(uint256 amount) transaction.
// Returns the remaining outstanding debt and the updated collateral ratio (×1e18).
// Returns (0, 0) if the position was fully closed after repayment.
func (p Precompile) repayDebt(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, errInvalidArgs(method.Name, 1, len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok || amount == nil {
		return nil, errInvalidArgType(method.Name, "amount")
	}

	owner := callerToCosmosAddr(contract.Caller())
	if err := p.collateralKeeper.RepayDebt(ctx, owner, math.NewIntFromBigInt(amount)); err != nil {
		return nil, err
	}

	// If the position was fully closed (zero collateral + zero debt), GetPosition
	// returns an error. Return zeros to indicate a closed position to the caller.
	pos, err := p.collateralKeeper.GetPosition(ctx, owner.String())
	var remainingDebt *big.Int
	var ratioBig *big.Int
	if err != nil {
		remainingDebt = big.NewInt(0)
		ratioBig = big.NewInt(0)
	} else {
		debtInt, _ := math.NewIntFromString(pos.DebtAmount)
		remainingDebt = debtInt.BigInt()
		ratio, _ := p.collateralKeeper.GetCollateralRatio(ctx, pos)
		ratioBig = ratio.MulInt64(1e18).TruncateInt().BigInt()
	}

	return method.Outputs.Pack(remainingDebt, ratioBig)
}

// liquidatePosition handles the liquidatePosition(address owner) transaction.
// The EVM caller acts as the liquidator; the owner argument identifies whose
// vault to liquidate. Both addresses are converted from EVM to Cosmos format.
func (p Precompile) liquidatePosition(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, errInvalidArgs(method.Name, 1, len(args))
	}

	ownerEVM, ok := args[0].(common.Address)
	if !ok {
		return nil, errInvalidArgType(method.Name, "owner")
	}

	owner := callerToCosmosAddr(ownerEVM)
	liquidator := callerToCosmosAddr(contract.Caller())

	if err := p.collateralKeeper.LiquidatePosition(ctx, owner, liquidator); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
}

// claimRewards handles the claimRewards() transaction.
// Returns the amount of uvusd claimed and credited to the caller's wallet.
func (p Precompile) claimRewards(
	ctx sdk.Context,
	contract *vm.Contract,
	_ *statedb.StateDB,
	method *abi.Method,
	_ []interface{},
) ([]byte, error) {
	owner := callerToCosmosAddr(contract.Caller())

	claimed, err := p.rewardsKeeper.ClaimRewards(ctx, owner)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(claimed.BigInt())
}
