package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance
func NewParams(stableDenom string) Params {
	return Params{
		StableDenom: stableDenom,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		StableDenom: "uvusd",
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.StableDenom == "" {
		return fmt.Errorf("stable_denom cannot be empty")
	}

	if err := sdk.ValidateDenom(p.StableDenom); err != nil {
		return fmt.Errorf("invalid stable_denom: %w", err)
	}

	return nil
}
