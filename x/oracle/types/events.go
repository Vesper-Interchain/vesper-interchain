package types

// Event types for the oracle module
const (
	EventTypePriceUpdate = "price_update"
	
	// Event attributes
	AttributeKeyDenom     = "denom"
	AttributeKeyPrice     = "price"
	AttributeKeySource    = "source"
	AttributeKeyTimestamp = "timestamp"
)