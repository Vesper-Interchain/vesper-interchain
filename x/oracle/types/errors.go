package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Sentinel errors for the oracle module.
// Each error is registered with a unique code so callers can programmatically
// distinguish between failure modes without parsing the error string.
var (
	// ErrPriceNotFound is returned when a price query references a denom that has
	// never had a price submitted, or whose entry was removed.
	ErrPriceNotFound = errorsmod.Register(ModuleName, 1, "price not found for denomination")

	// ErrInvalidPrice is returned when the submitted price string cannot be parsed
	// as a decimal or is negative.
	ErrInvalidPrice = errorsmod.Register(ModuleName, 2, "invalid price value")

	// ErrPriceStale is returned when the stored price is older than the configured
	// oracle_price_stale_seconds threshold.
	ErrPriceStale = errorsmod.Register(ModuleName, 3, "price data is stale")

	// ErrUnauthorizedOracle is returned when a MsgUpdatePrice is signed by an address
	// that does not match Params.OracleAddress.
	ErrUnauthorizedOracle = errorsmod.Register(ModuleName, 4, "unauthorized oracle account")

	// ErrEmptyDenom is returned when the denom field of a price message is empty.
	ErrEmptyDenom = errorsmod.Register(ModuleName, 5, "denom cannot be empty")

	// ErrEmptySource is returned when the source field of a price message is empty.
	// Requiring a source makes it easier to attribute prices for debugging.
	ErrEmptySource = errorsmod.Register(ModuleName, 6, "price source cannot be empty")

	// ErrInvalidSigner is returned by UpdateParams when the signer is not the
	// module's governance authority.
	ErrInvalidSigner = errorsmod.Register(ModuleName, 7, "invalid authority signer")
)
