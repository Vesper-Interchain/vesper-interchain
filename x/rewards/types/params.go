package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Params holds the governance-controlled parameters for the rewards module.
type Params struct {
	// RewardRate is the total number of uvusd tokens emitted per block across
	// all collateral depositors. The default of 1,000,000 equals 1 VUSD/block.
	// Individual allocations are proportional to each user's share of total collateral.
	RewardRate string `json:"reward_rate"`

	// StableDenom is the Cosmos denom of the stablecoin used for reward distribution.
	// Must match the stablecoin module's configured denom.
	StableDenom string `json:"stable_denom"`
}

// DefaultParams returns the canonical default parameters for a fresh chain.
func DefaultParams() Params {
	return Params{
		RewardRate:  "1000000", // 1 VUSD per block
		StableDenom: "uvusd",
	}
}

// Validate checks that all parameter values satisfy their invariants.
func (p Params) Validate() error {
	rate, ok := math.NewIntFromString(p.RewardRate)
	if !ok {
		return fmt.Errorf("invalid reward_rate: %s", p.RewardRate)
	}
	if rate.IsNegative() {
		return fmt.Errorf("reward_rate must be non-negative")
	}
	if p.StableDenom == "" {
		return fmt.Errorf("stable_denom cannot be empty")
	}
	return nil
}

// GetRewardRateAsInt parses and returns RewardRate as a math.Int.
func (p Params) GetRewardRateAsInt() (math.Int, error) {
	rate, ok := math.NewIntFromString(p.RewardRate)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid reward_rate: %s", p.RewardRate)
	}
	return rate, nil
}
