package types

// Event type and attribute constants for the rewards module.
// These events are emitted so off-chain indexers can track reward accrual
// and distribution without scanning raw state diffs.
const (
	// EventTypeClaimRewards is emitted when a user successfully claims their
	// pending uvusd rewards.
	EventTypeClaimRewards = "claim_rewards"

	// EventTypeUpdateShares is emitted each time a user's share weight is
	// updated (on deposit, withdrawal, or liquidation).
	EventTypeUpdateShares = "update_shares"

	// AttributeKeyOwner is the bech32 address of the account whose shares
	// or rewards were affected.
	AttributeKeyOwner = "owner"

	// AttributeKeyAmount is the uvusd amount claimed or minted.
	AttributeKeyAmount = "amount"

	// AttributeKeyShares is the updated share count for the owner.
	AttributeKeyShares = "shares"
)
