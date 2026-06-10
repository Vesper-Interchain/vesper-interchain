package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidSigner = errorsmod.Register(ModuleName, 1, "invalid signer")

	ErrInvalidAmount = errorsmod.Register(ModuleName, 2, "invalid amount")

	ErrInvalidDenom = errorsmod.Register(ModuleName, 3, "invalid denomination")

	ErrUnauthorized = errorsmod.Register(ModuleName, 4, "unauthorized")

	ErrSupplyUnderflow = errorsmod.Register(ModuleName, 5, "burn amount exceeds minted supply")
)
