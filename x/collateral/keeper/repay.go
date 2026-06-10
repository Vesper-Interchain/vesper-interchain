package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) RepayDebt(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	currentDebt, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return types.ErrInvalidAmount
	}

	if amount.GT(currentDebt) {
		amount = currentDebt
	}
	newDebt := currentDebt.Sub(amount)

	if err := k.stablecoinKeeper.BurnStablecoin(ctx, owner, amount); err != nil {
		return err
	}

	position.DebtAmount = newDebt.String()
	position.UpdatedAt = ctx.BlockTime().Unix()

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
