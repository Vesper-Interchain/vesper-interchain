package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/types"
)

// EnqueuePosition adds an unhealthy vault to the liquidation queue and returns
// the assigned position_id. Any account may call this; the function verifies
// the position's health before accepting it to prevent spam entries for healthy vaults.
//
// The snapshot of collateral and debt stored in the queue entry reflects the
// state at enqueue time. By the time the liquidation is executed the live position
// values may differ (e.g. due to price changes), so the executor always reads
// live state from the collateral keeper rather than the queue snapshot.
func (k Keeper) EnqueuePosition(ctx sdk.Context, owner sdk.AccAddress) (uint64, error) {
	position, err := k.collateralKeeper.GetPosition(ctx, owner.String())
	if err != nil {
		return 0, fmt.Errorf("position not found for owner %s: %w", owner, err)
	}

	isHealthy, err := k.collateralKeeper.IsPositionHealthy(ctx, position)
	if err != nil {
		return 0, err
	}
	if isHealthy {
		return 0, types.ErrLiquidationFailed.Wrap("position is healthy, cannot queue for liquidation")
	}

	// Assign a unique monotonic ID for this queue entry.
	id, err := k.PositionCounter.Next(ctx)
	if err != nil {
		return 0, err
	}

	entry := types.LiquidationQueue{
		PositionId:       id,
		Owner:            owner.String(),
		CollateralType:   position.CollateralDenom,
		CollateralAmount: position.CollateralAmount,
		DebtAmount:       position.DebtAmount,
		Timestamp:        ctx.BlockTime().Unix(),
		Status:           "pending",
	}

	if err := k.Queue.Set(ctx, id, entry); err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeQueueLiquidation,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyPositionID, fmt.Sprintf("%d", id)),
		),
	)

	return id, nil
}

// GetQueueEntry retrieves a liquidation queue entry by its position_id.
// Returns ErrPositionNotFound if the ID does not exist in the queue (e.g. already
// executed or never enqueued).
func (k Keeper) GetQueueEntry(ctx sdk.Context, positionID uint64) (types.LiquidationQueue, error) {
	entry, err := k.Queue.Get(ctx, positionID)
	if err != nil {
		return types.LiquidationQueue{}, types.ErrPositionNotFound
	}
	return entry, nil
}

// RemoveQueueEntry deletes a queue entry after it has been executed or cancelled.
// Keeping the entry around after execution would allow double-execution, so
// this must always be called after a successful ExecuteLiquidation.
func (k Keeper) RemoveQueueEntry(ctx sdk.Context, positionID uint64) error {
	return k.Queue.Remove(ctx, positionID)
}
