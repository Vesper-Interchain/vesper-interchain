package types

// Event type and attribute constants for the liquidation module.
// These events are emitted on queue and execution so off-chain liquidation
// bots can monitor for profitable opportunities without polling state directly.
const (
	// EventTypeLiquidatePosition is emitted when a queued liquidation is successfully
	// executed (collateral seized, debt burned).
	EventTypeLiquidatePosition = "liquidate_position"

	// EventTypeQueueLiquidation is emitted when an unhealthy position is added
	// to the pending liquidation queue.
	EventTypeQueueLiquidation = "queue_liquidation"

	// AttributeKeyOwner is the bech32 address of the vault owner being liquidated.
	AttributeKeyOwner = "owner"

	// AttributeKeyLiquidator is the bech32 address of the account executing the liquidation.
	AttributeKeyLiquidator = "liquidator"

	// AttributeKeyPenalty is the liquidation penalty amount in USD included in the seized collateral.
	AttributeKeyPenalty = "penalty"

	// AttributeKeyPositionID is the uint64 queue position ID assigned at enqueue time.
	AttributeKeyPositionID = "position_id"
)
