package types

// Event type and attribute key constants for the oracle module.
// These strings are emitted as ABCI events so off-chain services (indexers,
// monitoring dashboards, oracle feeders) can subscribe to price activity
// without parsing transaction bytes directly.
const (
	// EventTypePriceUpdate is emitted each time a new price is successfully stored.
	EventTypePriceUpdate = "price_update"

	// AttributeKeyDenom identifies the denomination whose price was updated.
	AttributeKeyDenom = "denom"

	// AttributeKeyPrice is the submitted price value as a decimal string.
	AttributeKeyPrice = "price"

	// AttributeKeySource identifies the data source reported by the oracle feeder.
	AttributeKeySource = "source"

	// AttributeKeyTimestamp is the Unix block timestamp at the time of submission.
	AttributeKeyTimestamp = "timestamp"
)
