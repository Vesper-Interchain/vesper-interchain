package types

import (
	"cosmossdk.io/errors"
)

// Sentinel errors for the liquidation module.
var (
	// ErrInvalidSigner is returned when a governance-only message is signed by
	// an account that is not the module's designated authority (x/gov).
	ErrInvalidSigner = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")

	// ErrLiquidationFailed is returned when a liquidation attempt targets a
	// position that is still healthy (collateral ratio above the liquidation
	// threshold) or has no debt.
	ErrLiquidationFailed = errors.Register(ModuleName, 1101, "liquidation failed")

	// ErrInsufficientLiquidatorBalance is returned when the liquidator does not
	// hold enough uvusd to cover the full outstanding debt of the target position.
	ErrInsufficientLiquidatorBalance = errors.Register(ModuleName, 1102, "liquidator has insufficient balance to repay debt")

	// ErrPositionNotFound is returned when an ExecuteLiquidation call references
	// a position_id that does not exist in the queue (never enqueued, or already executed).
	ErrPositionNotFound = errors.Register(ModuleName, 1103, "position not found in liquidation queue")
)
