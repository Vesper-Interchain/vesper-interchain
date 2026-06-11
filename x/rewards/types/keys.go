// Package types contains all type definitions, constants, error codes, events,
// and keeper interfaces for the rewards module. The rewards module distributes
// uvusd stablecoin rewards to collateral depositors every block using the
// MasterChef accumulator pattern (proportional to deposited collateral).
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name of the rewards module, used for the
	// module account, store key, and event attribute namespacing.
	ModuleName = "rewards"

	// StoreKey is the KV-store key registered for the rewards module.
	StoreKey = ModuleName
)

var (
	// ParamsKey is the collections prefix for JSON-encoded module parameters.
	ParamsKey = collections.NewPrefix("p_rewards")

	// RewardAccumulatorKey is the collections prefix for the single global
	// accumulator item (a LegacyDec stored as a string).
	RewardAccumulatorKey = collections.NewPrefix("reward_acc")

	// TotalSharesKey is the collections prefix for the total depositor shares
	// item (an Int stored as a string).
	TotalSharesKey = collections.NewPrefix("total_shares")

	// UserSharesPrefix is the collections prefix for the per-user shares map
	// (owner bech32 string → Int string).
	UserSharesPrefix = collections.NewPrefix("user_shares")

	// UserRewardDebtPrefix is the collections prefix for the per-user reward
	// debt map (owner bech32 string → LegacyDec string).
	UserRewardDebtPrefix = collections.NewPrefix("user_debt")

	// PendingRewardsPrefix is reserved for a future per-user pending rewards
	// cache; not currently used in the keeper implementation.
	PendingRewardsPrefix = collections.NewPrefix("pending_rewards")
)
