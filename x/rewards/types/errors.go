package types

import "cosmossdk.io/errors"

// Sentinel errors for the rewards module.
var (
	// ErrNoRewardsToClaim is returned by ClaimRewards when the calculated pending
	// reward amount is zero (the user either has no shares or the accumulator has
	// not advanced since their last claim).
	ErrNoRewardsToClaim = errors.Register(ModuleName, 1200, "no rewards to claim")

	// ErrInvalidAmount is returned when an amount parameter fails basic validation
	// (zero, negative, or unparseable).
	ErrInvalidAmount = errors.Register(ModuleName, 1201, "invalid amount")
)
