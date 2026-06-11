package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Sentinel errors for the stablecoin module.
var (
	// ErrInvalidSigner is returned when a governance message is signed by an
	// address that is not the module's designated authority.
	ErrInvalidSigner = errorsmod.Register(ModuleName, 1, "invalid signer")

	// ErrInvalidAmount is returned when an amount is zero, negative, or cannot
	// be parsed as a valid integer.
	ErrInvalidAmount = errorsmod.Register(ModuleName, 2, "invalid amount")

	// ErrInvalidDenom is returned when a coin denom does not match the configured
	// Params.StableDenom (normally "uvusd").
	ErrInvalidDenom = errorsmod.Register(ModuleName, 3, "invalid denomination")

	// ErrUnauthorized is returned for operations that require a privileged signer
	// but receive an unprivileged one.
	ErrUnauthorized = errorsmod.Register(ModuleName, 4, "unauthorized")

	// ErrSupplyUnderflow is returned when a burn request would push the cumulative
	// burned counter above the cumulative minted counter — an invariant that must
	// never be violated.
	ErrSupplyUnderflow = errorsmod.Register(ModuleName, 5, "burn amount exceeds minted supply")
)
