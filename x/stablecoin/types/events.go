package types

const (
	// EventTypeMint is emitted when stablecoins are minted
	EventTypeMint = "mint"
	
	// EventTypeBurn is emitted when stablecoins are burned
	EventTypeBurn = "burn"
	
	// EventTypeUpdateParams is emitted when module parameters are updated
	EventTypeUpdateParams = "update_params"

	// Event attributes
	AttributeKeyRecipient = "recipient"
	AttributeKeySender    = "sender"
	AttributeKeyAmount    = "amount"
)
