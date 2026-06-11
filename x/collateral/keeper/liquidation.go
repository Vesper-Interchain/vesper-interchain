package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// LiquidatePosition allows any account (the "liquidator") to close an
// under-collateralised vault by repaying its debt in exchange for a discounted
// share of the collateral.
//
// Liquidation flow:
//  1. Verify the oracle price is fresh — stale prices must not drive liquidations.
//  2. Confirm the position's collateral ratio is below the liquidation ratio.
//  3. Calculate the debt to repay and the collateral to award (debt + penalty).
//  4. The liquidator burns the required uvusd (via stablecoin keeper).
//  5. The corresponding collateral is released from module escrow to the liquidator.
//  6. The position is updated or deleted.
//
// Note on the burn call: BurnStablecoin internally sends tokens from the liquidator
// to the module account and then burns them in a single operation. There must be no
// separate SendCoinsFromAccountToModule call before it to avoid a double-spend.
func (k Keeper) LiquidatePosition(ctx sdk.Context, owner sdk.AccAddress, liquidator sdk.AccAddress) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Require a fresh oracle price to prevent liquidation under artificially stale data.
	if err := k.CheckOraclePrice(ctx, params.SupportedCollateralDenom); err != nil {
		return err
	}

	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	// Only under-collateralised positions (CR < liquidation ratio) may be liquidated.
	isHealthy, err := k.IsPositionHealthy(ctx, position)
	if err != nil {
		return err
	}
	if isHealthy {
		return types.ErrLiquidationFailed
	}

	// A position with no debt cannot be liquidated — nothing to repay.
	if position.DebtAmount == "0" {
		return types.ErrLiquidationFailed
	}

	// Determine the exact collateral to seize and debt to repay, including the
	// liquidation penalty bonus that incentivises liquidators.
	collateralToGive, penaltyUSD, debtToRepayUVUSD, err := k.CalculateLiquidationOutput(ctx, position)
	if err != nil {
		return err
	}

	// Burn the required uvusd from the liquidator's wallet.
	// BurnStablecoin handles the send-to-module + burn sequence internally.
	if err := k.stablecoinKeeper.BurnStablecoin(ctx, liquidator, debtToRepayUVUSD); err != nil {
		return fmt.Errorf("%w: %v", types.ErrInsufficientLiquidatorBalance, err)
	}

	// Transfer the seized collateral (including the penalty bonus) to the liquidator.
	collateralCoins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, collateralToGive))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidator, collateralCoins); err != nil {
		return err
	}

	// Recalculate remaining balances after the liquidation.
	remainingCollateral, _ := math.NewIntFromString(position.CollateralAmount)
	newCollateral := remainingCollateral.Sub(collateralToGive)
	newDebt, _ := math.NewIntFromString(position.DebtAmount)
	newDebt = newDebt.Sub(debtToRepayUVUSD)

	if newCollateral.IsZero() && newDebt.IsZero() {
		// The entire position was consumed by the liquidation; delete it.
		if err := k.DeletePosition(ctx, owner.String()); err != nil {
			return err
		}
	} else {
		// Partial liquidation: update the position with the remaining balances.
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
