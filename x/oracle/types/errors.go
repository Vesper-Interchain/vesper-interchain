package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/oracle module sentinel errors
var (
	ErrPriceNotFound      = errorsmod.Register(ModuleName, 1, "price not found for denomination")
	ErrInvalidPrice       = errorsmod.Register(ModuleName, 2, "invalid price value")
	ErrPriceStale         = errorsmod.Register(ModuleName, 3, "price data is stale")
	ErrUnauthorizedOracle = errorsmod.Register(ModuleName, 4, "unauthorized oracle account")
	ErrEmptyDenom         = errorsmod.Register(ModuleName, 5, "denom cannot be empty")
	ErrEmptySource        = errorsmod.Register(ModuleName, 6, "price source cannot be empty")

	// generated UpdateParams uses this
	ErrInvalidSigner = errorsmod.Register(ModuleName, 7, "invalid authority signer")
)