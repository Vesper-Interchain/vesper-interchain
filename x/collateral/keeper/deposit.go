package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) DepositCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	minAmount, ok := math.NewIntFromString(params.MinCollateralAmount)
	if !ok {
		return types.ErrInvalidAmount
	}
	if amount.LT(minAmount) {
		return types.ErrCollateralTooLow
	}

	coins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	position, err := k.GetPosition(ctx, owner.String())
	if err != nil && err != types.ErrPositionNotFound {
		return err
	}

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
	} else {
		currentAmount, _ := math.NewIntFromString(position.CollateralAmount)
		position.CollateralAmount = currentAmount.Add(amount).String()
		position.UpdatedAt = now
	}

	if err := k.SetPosition(ctx, position); err != nil {
		return err
	}

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
