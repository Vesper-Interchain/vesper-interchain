package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	if err := k.CheckOraclePrice(ctx, params.SupportedCollateralDenom); err != nil {
		return err
	}

	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	currentCollateral, ok := math.NewIntFromString(position.CollateralAmount)
	if !ok {
		return types.ErrInvalidAmount
	}
	if amount.GT(currentCollateral) {
		return types.ErrInsufficientCollateral
	}

	newCollateral := currentCollateral.Sub(amount)
	currentDebt, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return types.ErrInvalidAmount
	}

	if currentDebt.IsPositive() {
		tempPosition := position
		tempPosition.CollateralAmount = newCollateral.String()
		isHealthy, err := k.IsPositionHealthy(ctx, tempPosition)
		if err != nil {
			return err
		}
		if !isHealthy {
			return types.ErrCollateralRatioTooLow
		}
	}

	coins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, coins); err != nil {
		return err
	}

	if newCollateral.IsZero() && currentDebt.IsZero() {
		if err := k.DeletePosition(ctx, owner.String()); err != nil {
			return err
		}
	} else {
		position.CollateralAmount = newCollateral.String()
		position.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.SetPosition(ctx, position); err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCollateral,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}
