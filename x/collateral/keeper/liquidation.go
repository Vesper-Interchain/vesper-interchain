package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) LiquidatePosition(ctx sdk.Context, owner sdk.AccAddress, liquidator sdk.AccAddress) error {
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

	isHealthy, err := k.IsPositionHealthy(ctx, position)
	if err != nil {
		return err
	}
	if isHealthy {
		return types.ErrLiquidationFailed
	}

	if position.DebtAmount == "0" {
		return types.ErrLiquidationFailed
	}

	collateralToGive, penaltyUSD, debtToRepayUVUSD, err := k.CalculateLiquidationOutput(ctx, position)
	if err != nil {
		return err
	}

	// Liquidator pays the debt
	debtCoins := sdk.NewCoins(sdk.NewCoin("uvusd", debtToRepayUVUSD))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidator, types.ModuleName, debtCoins); err != nil {
		return fmt.Errorf("%w: %v", types.ErrInsufficientLiquidatorBalance, err)
	}

	// Burn the uvUSD
	if err := k.stablecoinKeeper.BurnStablecoin(ctx, liquidator, debtToRepayUVUSD); err != nil {
		return err
	}

	// Send collateral to liquidator
	collateralCoins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, collateralToGive))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidator, collateralCoins); err != nil {
		return err
	}

	// Update position
	remainingCollateral, _ := math.NewIntFromString(position.CollateralAmount)
	newCollateral := remainingCollateral.Sub(collateralToGive)
	newDebt, _ := math.NewIntFromString(position.DebtAmount)
	newDebt = newDebt.Sub(debtToRepayUVUSD)

	if newCollateral.IsZero() && newDebt.IsZero() {
		if err := k.DeletePosition(ctx, owner.String()); err != nil {
			return err
		}
	} else {
		position.CollateralAmount = newCollateral.String()
		position.DebtAmount = newDebt.String()
		position.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.SetPosition(ctx, position); err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidatePosition,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyLiquidator, liquidator.String()),
			sdk.NewAttribute("collateral_seized", collateralToGive.String()),
			sdk.NewAttribute("debt_repaid", debtToRepayUVUSD.String()),
			sdk.NewAttribute(types.AttributeKeyPenalty, penaltyUSD.String()),
		),
	)

	return nil
}
