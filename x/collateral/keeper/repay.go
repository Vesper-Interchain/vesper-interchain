package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// RepayDebt burns uvusd tokens from the owner's wallet to reduce or eliminate
// their vault's outstanding debt.
//
// Rules enforced:
//   - If amount exceeds the current debt the repayment is capped to the exact
//     debt balance, preventing over-repayment errors.
//   - The stablecoin keeper handles the send-to-module and burn in one atomic step.
//   - When both debt and collateral reach zero after repayment the position
//     record is deleted from state (full vault close).
//
// No oracle price check is required here because repayment can only improve
// the health of the position.
func (k Keeper) RepayDebt(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	currentDebt, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return types.ErrInvalidAmount
	}

	// Cap repayment at the outstanding balance so users do not need to query
	// the exact debt amount before calling repay.
	if amount.GT(currentDebt) {
		amount = currentDebt
	}
	newDebt := currentDebt.Sub(amount)

	// Burn exactly `amount` uvusd from the owner's account via the stablecoin keeper.
	if err := k.stablecoinKeeper.BurnStablecoin(ctx, owner, amount); err != nil {
		return err
	}

	position.DebtAmount = newDebt.String()
	position.UpdatedAt = ctx.BlockTime().Unix()

	// If the debt is fully cleared, check whether collateral is also zero.
	// If so, remove the position entirely to free up storage.
	if newDebt.IsZero() {
		collateralAmount, _ := math.NewIntFromString(position.CollateralAmount)
		if collateralAmount.IsZero() {
			if err := k.DeletePosition(ctx, owner.String()); err != nil {
				return err
			}
			return nil
		}
	}

	if err := k.SetPosition(ctx, position); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRepayDebt,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDebt, position.DebtAmount),
		),
	)

	return nil
}
