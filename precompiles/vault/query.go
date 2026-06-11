package vault

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getPosition handles the getPosition(address owner) view call.
// Returns (collateralAmount, debtAmount, collateralRatio) for the given owner.
// All amounts are in their smallest denomination (uatom, uvusd).
// The collateral ratio is scaled by 1e18 to match Solidity's uint256 convention
// for representing decimals. Returns (0, 0, 0) if the address has no open position.
func (p Precompile) getPosition(
	ctx sdk.Context,
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
	pos, err := p.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		// No position exists for this address — return all zeros.
		zero := big.NewInt(0)
		return method.Outputs.Pack(zero, zero, zero)
	}

	collateralAmt, _ := math.NewIntFromString(pos.CollateralAmount)
	debtAmt, _ := math.NewIntFromString(pos.DebtAmount)

	var ratioBig *big.Int
	if debtAmt.IsZero() {
		// No debt means infinite collateral ratio; return 0 to avoid confusion
		// in downstream Solidity code that would divide by a sentinel value.
		ratioBig = big.NewInt(0)
	} else {
		ratio, err := p.collateralKeeper.GetCollateralRatio(ctx, pos)
		if err != nil {
			ratioBig = big.NewInt(0)
		} else {
			ratioBig = ratio.MulInt64(1e18).TruncateInt().BigInt()
		}
	}

	return method.Outputs.Pack(collateralAmt.BigInt(), debtAmt.BigInt(), ratioBig)
}

// getPendingRewards handles the getPendingRewards(address owner) view call.
// Returns the amount of uvusd that the given owner can claim without modifying state.
// Used by front-end UIs and the VesperVault wrapper contract to display claimable rewards.
func (p Precompile) getPendingRewards(
	ctx sdk.Context,
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
	pending := p.rewardsKeeper.GetPendingRewards(ctx, owner)
	return method.Outputs.Pack(pending.BigInt())
}
