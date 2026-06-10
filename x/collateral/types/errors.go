package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrPositionNotFound        = errorsmod.Register(ModuleName, 1, "position not found")
	ErrInsufficientCollateral  = errorsmod.Register(ModuleName, 2, "insufficient collateral")
	ErrCollateralRatioTooLow   = errorsmod.Register(ModuleName, 3, "collateral ratio too low")
	ErrInvalidAmount           = errorsmod.Register(ModuleName, 4, "invalid amount")
	ErrUnauthorized            = errorsmod.Register(ModuleName, 5, "unauthorized")
	ErrDebtExists              = errorsmod.Register(ModuleName, 6, "debt exists, cannot withdraw all collateral")
	ErrLiquidationFailed       = errorsmod.Register(ModuleName, 7, "liquidation failed: position is healthy")
	ErrInvalidCollateralDenom  = errorsmod.Register(ModuleName, 8, "invalid collateral denom")
	ErrCollateralTooLow        = errorsmod.Register(ModuleName, 9, "collateral amount below minimum")
	ErrMintLimitExceeded       = errorsmod.Register(ModuleName, 10, "mint amount exceeds limit")
	ErrDebtTooLow              = errorsmod.Register(ModuleName, 11, "debt amount below minimum")
	ErrInsufficientLiquidatorBalance = errorsmod.Register(ModuleName, 12, "liquidator doesn't have enough uvUSD")
)
