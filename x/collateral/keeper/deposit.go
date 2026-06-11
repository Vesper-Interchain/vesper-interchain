package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// DepositCollateral moves uatom from the depositor's account into the collateral
// module account and updates (or creates) the user's Position.
//
// Rules enforced:
//   - The deposit amount must be at least Params.MinCollateralAmount.
//   - The deposited denom is always Params.SupportedCollateralDenom (uatom).
//   - Subsequent deposits accumulate on an existing position rather than opening a new one.
//
// After the position is written, the rewards module is notified of the updated
// share count so that reward accrual reflects the new collateral balance.
func (k Keeper) DepositCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Enforce minimum deposit to prevent dust positions that are expensive to liquidate.
	minAmount, ok := math.NewIntFromString(params.MinCollateralAmount)
	if !ok {
		return types.ErrInvalidAmount
	}
	if amount.LT(minAmount) {
		return types.ErrCollateralTooLow
	}

	// Transfer collateral from the depositor to the module escrow account.
	coins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	// Load existing position, or build a new one if this is the first deposit.
	position, err := k.GetPosition(ctx, owner.String())
	now := ctx.BlockTime().Unix()

	if err == types.ErrPositionNotFound {
		position = types.Position{
			Owner:            owner.String(),
			CollateralDenom:  params.SupportedCollateralDenom,
			CollateralAmount: amount.String(),
			DebtAmount:       "0",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	} else if err != nil {
		return err
	} else {
		// Existing position: add to the existing collateral balance.
		currentAmount, _ := math.NewIntFromString(position.CollateralAmount)
		position.CollateralAmount = currentAmount.Add(amount).String()
		position.UpdatedAt = now
	}

	if err := k.SetPosition(ctx, position); err != nil {
		return err
	}

	// Inform the rewards module of the new share count (equal to collateral in uatom).
	newShares, _ := math.NewIntFromString(position.CollateralAmount)
	k.notifyRewards(ctx, owner, newShares)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDepositCollateral,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyCollateral, position.CollateralAmount),
		),
	)

	return nil
}
