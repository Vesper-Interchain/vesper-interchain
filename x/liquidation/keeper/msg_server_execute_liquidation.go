package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	colltypes "github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/types"
)

func (k msgServer) ExecuteLiquidation(goCtx context.Context, msg *types.MsgExecuteLiquidation) (*types.MsgExecuteLiquidationResponse, error) {
	liquidator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid liquidator address")
	}

	// TODO: In a complete implementation with position IDs, we would:
	// 1. Look up the position by position_id from a position registry
	// 2. Get the owner from the position
	// 3. Call LiquidatePosition(ctx, liquidator, owner)
	// 
	// For now, this is a placeholder that demonstrates the flow
	_ = liquidator // TODO: Remove when position lookup is implemented

	return &types.MsgExecuteLiquidationResponse{}, nil
}

// LiquidatePosition handles liquidation of a specific position
// This is called by the liquidation module to execute liquidations
func (k Keeper) LiquidatePosition(ctx sdk.Context, liquidator sdk.AccAddress, owner sdk.AccAddress) error {
	collateralParams, err := k.collateralKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	// Check oracle price freshness
	if err := k.checkOraclePriceStale(ctx, collateralParams); err != nil {
		return err
	}

	position, err := k.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	// Verify position is unhealthy
	isHealthy, err := k.collateralKeeper.IsPositionHealthy(ctx, position)
	if err != nil {
		return err
	}
	if isHealthy {
		return types.ErrLiquidationFailed.Wrapf("position is still healthy, cannot liquidate")
	}

	if position.DebtAmount == "0" {
		return types.ErrLiquidationFailed.Wrapf("position has no debt")
	}

	// Calculate liquidation outputs
	collateralToGive, penaltyUSD, debtToRepayUVUSD, err := k.collateralKeeper.CalculateLiquidationOutput(ctx, position)
	if err != nil {
		return err
	}

	// Liquidator pays the debt
	debtCoins := sdk.NewCoins(sdk.NewCoin("uvusd", debtToRepayUVUSD))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidator, colltypes.ModuleName, debtCoins); err != nil {
		return fmt.Errorf("%w: %v", types.ErrInsufficientLiquidatorBalance, err)
	}

	// Burn the uvUSD
	if err := k.stablecoinKeeper.BurnStablecoin(ctx, liquidator, debtToRepayUVUSD); err != nil {
		return err
	}

	// Send collateral to liquidator
	collateralCoins := sdk.NewCoins(sdk.NewCoin(collateralParams.SupportedCollateralDenom, collateralToGive))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, colltypes.ModuleName, liquidator, collateralCoins); err != nil {
		return err
	}

	// Update position
	remainingCollateral, _ := math.NewIntFromString(position.CollateralAmount)
	newCollateral := remainingCollateral.Sub(collateralToGive)
	newDebt, _ := math.NewIntFromString(position.DebtAmount)
	newDebt = newDebt.Sub(debtToRepayUVUSD)

	if newCollateral.IsZero() && newDebt.IsZero() {
		if err := k.collateralKeeper.DeletePosition(ctx, owner.String()); err != nil {
			return err
		}
	} else {
		position.CollateralAmount = newCollateral.String()
		position.DebtAmount = newDebt.String()
		position.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.collateralKeeper.SetPosition(ctx, position); err != nil {
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

// checkOraclePriceStale checks if oracle price is stale
func (k Keeper) checkOraclePriceStale(ctx sdk.Context, params colltypes.Params) error {
	stale, err := k.oracleKeeper.IsPriceStale(ctx, params.SupportedCollateralDenom, params.OraclePriceStaleSeconds)
	if err != nil {
		return err
	}
	if stale {
		return fmt.Errorf("oracle price for %s is stale", params.SupportedCollateralDenom)
	}
	return nil
}
