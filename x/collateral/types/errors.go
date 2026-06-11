package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Sentinel errors for the collateral module.
var (
	// ErrPositionNotFound is returned when a vault operation references an owner
	// address that has no open position.
	ErrPositionNotFound = errorsmod.Register(ModuleName, 1, "position not found")

	// ErrInsufficientCollateral is returned when a withdrawal or transfer would
	// reduce collateral below the requested amount.
	ErrInsufficientCollateral = errorsmod.Register(ModuleName, 2, "insufficient collateral")

	// ErrCollateralRatioTooLow is returned when a requested operation (withdrawal
	// or additional minting) would push the collateral ratio below the liquidation threshold.
	ErrCollateralRatioTooLow = errorsmod.Register(ModuleName, 3, "collateral ratio too low")

	// ErrInvalidAmount is returned when an amount field cannot be parsed as a
	// positive integer, or when a required amount is zero or negative.
	ErrInvalidAmount = errorsmod.Register(ModuleName, 4, "invalid amount")

	// ErrUnauthorized is returned for operations that require a specific signer
	// (e.g. governance) but receive a different address.
	ErrUnauthorized = errorsmod.Register(ModuleName, 5, "unauthorized")

	// ErrDebtExists is returned when a user attempts to withdraw all collateral
	// while an outstanding debt balance remains.
	ErrDebtExists = errorsmod.Register(ModuleName, 6, "debt exists, cannot withdraw all collateral")

	// ErrLiquidationFailed is returned when a liquidation attempt is made against
	// a position that is still above the liquidation ratio (i.e. healthy).
	ErrLiquidationFailed = errorsmod.Register(ModuleName, 7, "liquidation failed: position is healthy")

	// ErrInvalidCollateralDenom is returned when a deposit specifies a token denom
	// that is not listed in Params.SupportedCollateralDenom.
	ErrInvalidCollateralDenom = errorsmod.Register(ModuleName, 8, "invalid collateral denom")

	// ErrCollateralTooLow is returned when a deposit amount is below the configured
	// Params.MinCollateralAmount floor.
	ErrCollateralTooLow = errorsmod.Register(ModuleName, 9, "collateral amount below minimum")

	// ErrMintLimitExceeded is returned when the requested mint amount would push the
	// user's debt above MaxLTV * collateral_value.
	ErrMintLimitExceeded = errorsmod.Register(ModuleName, 10, "mint amount exceeds limit")

	// ErrDebtTooLow is returned when the resulting debt after a mint would be below
	// the minimum debt floor (Params.MinDebtAmount), making the position
	// uneconomical for future liquidators.
	ErrDebtTooLow = errorsmod.Register(ModuleName, 11, "debt amount below minimum")

	// ErrInsufficientLiquidatorBalance is returned when the liquidator does not hold
	// enough uvusd to cover the full outstanding debt of the position being liquidated.
	ErrInsufficientLiquidatorBalance = errorsmod.Register(ModuleName, 12, "liquidator doesn't have enough uvUSD")
)
