package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/liquidation module sentinel errors
var (
	ErrInvalidSigner                 = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrLiquidationFailed             = errors.Register(ModuleName, 1101, "liquidation failed")
	ErrInsufficientLiquidatorBalance = errors.Register(ModuleName, 1102, "liquidator has insufficient balance to repay debt")
)
