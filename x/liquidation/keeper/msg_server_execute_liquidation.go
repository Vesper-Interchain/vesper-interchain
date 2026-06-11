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

// ExecuteLiquidation processes a MsgExecuteLiquidation transaction.
// It reads the queued position, delegates the actual liquidation to LiquidatePosition,
// and then removes the queue entry to prevent double-execution.
//
// The liquidator (msg.Creator) must hold enough uvusd in their wallet to cover
// the full outstanding debt of the queued position at the time of execution.
func (k msgServer) ExecuteLiquidation(goCtx context.Context, msg *types.MsgExecuteLiquidation) (*types.MsgExecuteLiquidationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid liquidator address")
	}

	// Fetch the queue entry to obtain the owner address and verify the position exists.
	entry, err := k.Keeper.GetQueueEntry(ctx, msg.PositionId)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrPositionNotFound, "position_id %d not in queue", msg.PositionId)
	}

	owner, err := sdk.AccAddressFromBech32(entry.Owner)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid owner address in queue entry")
	}

	// Delegate the full liquidation logic to LiquidatePosition which handles the
	// health check, debt burn, collateral transfer, and position update/deletion.
	if err := k.Keeper.LiquidatePosition(ctx, liquidator, owner); err != nil {
		return nil, err
	}

	// Remove the entry from the queue so it cannot be executed again.
	if err := k.Keeper.RemoveQueueEntry(ctx, msg.PositionId); err != nil {
		return nil, err
	}

	return &types.MsgExecuteLiquidationResponse{}, nil
}

// LiquidatePosition executes the full liquidation of a position by the given liquidator.
// It is called both by ExecuteLiquidation (queue-based path) and can be called directly
// by the collateral module's direct-liquidation message handler.
//
// This function re-checks all preconditions (oracle freshness, position health) from
// live state rather than relying on the queue snapshot, because conditions may have
// changed between enqueue and execution (e.g. price recovery making the position healthy).
func (k Keeper) LiquidatePosition(ctx sdk.Context, liquidator sdk.AccAddress, owner sdk.AccAddress) error {
	collateralParams, err := k.collateralKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	// Verify oracle price is fresh before any liquidation calculation.
	if err := k.checkOraclePriceStale(ctx, collateralParams); err != nil {
		return err
	}

	position, err := k.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	// Re-check health from live state; the position may have recovered since it was queued.
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

	// Compute the collateral to award and the debt the liquidator must repay.
	collateralToGive, penaltyUSD, debtToRepayUVUSD, err := k.collateralKeeper.CalculateLiquidationOutput(ctx, position)
	if err != nil {
		return err
	}

	// Burn the debt from the liquidator's wallet. BurnStablecoin handles the
	// send-to-module + burn in one step; no separate SendCoinsFromAccountToModule call.
	if err := k.stablecoinKeeper.BurnStablecoin(ctx, liquidator, debtToRepayUVUSD); err != nil {
		return fmt.Errorf("%w: %v", types.ErrInsufficientLiquidatorBalance, err)
	}

	// Release the seized collateral (plus penalty) to the liquidator.
	collateralCoins := sdk.NewCoins(sdk.NewCoin(collateralParams.SupportedCollateralDenom, collateralToGive))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, colltypes.ModuleName, liquidator, collateralCoins); err != nil {
		return err
	}

	// Recalculate remaining balances and update or delete the position.
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

// checkOraclePriceStale is a helper that reads the oracle staleness window from
// collateral params and delegates the stale check to the oracle keeper.
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
