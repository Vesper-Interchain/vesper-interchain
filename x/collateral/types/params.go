package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewParams(
	liquidationRatio, liquidationPenalty, minCollateralAmount, supportedCollateralDenom,
	maxLoanToValue, minDebtAmount string, oraclePriceStaleSeconds int64,
) Params {
	return Params{
		LiquidationRatio:          liquidationRatio,
		LiquidationPenalty:        liquidationPenalty,
		MinCollateralAmount:       minCollateralAmount,
		SupportedCollateralDenom:  supportedCollateralDenom,
		MaxLoanToValue:            maxLoanToValue,
		MinDebtAmount:             minDebtAmount,
		OraclePriceStaleSeconds:   oraclePriceStaleSeconds,
	}
}

func DefaultParams() Params {
	return Params{
		LiquidationRatio:         "1.5",      // 150%
		LiquidationPenalty:       "0.05",     // 5%
		MinCollateralAmount:      "1000000",  // 1 ATOM (6 decimals)
		SupportedCollateralDenom: "uatom",
		MaxLoanToValue:           "0.8",      // 80%
		MinDebtAmount:            "1000000",  // 1 USD (6 decimals)
		OraclePriceStaleSeconds:  3600,       // 1 hour
	}
}

func (p Params) Validate() error {
	if p.LiquidationRatio == "" {
		return fmt.Errorf("liquidation_ratio cannot be empty")
	}
	if p.LiquidationPenalty == "" {
		return fmt.Errorf("liquidation_penalty cannot be empty")
	}
	if p.MinCollateralAmount == "" {
		return fmt.Errorf("min_collateral_amount cannot be empty")
	}
	if p.SupportedCollateralDenom == "" {
		return fmt.Errorf("supported_collateral_denom cannot be empty")
	}
	if p.MaxLoanToValue == "" {
		return fmt.Errorf("max_loan_to_value cannot be empty")
	}
	if p.MinDebtAmount == "" {
		return fmt.Errorf("min_debt_amount cannot be empty")
	}
	if p.OraclePriceStaleSeconds <= 0 {
		return fmt.Errorf("oracle_price_stale_seconds must be positive")
	}

	ratio, err := math.LegacyNewDecFromStr(p.LiquidationRatio)
	if err != nil {
		return fmt.Errorf("invalid liquidation_ratio: %w", err)
	}
	if ratio.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("liquidation_ratio must be positive")
	}

	penalty, err := math.LegacyNewDecFromStr(p.LiquidationPenalty)
	if err != nil {
		return fmt.Errorf("invalid liquidation_penalty: %w", err)
	}
	if penalty.IsNegative() {
		return fmt.Errorf("liquidation_penalty cannot be negative")
	}

	ltv, err := math.LegacyNewDecFromStr(p.MaxLoanToValue)
	if err != nil {
		return fmt.Errorf("invalid max_loan_to_value: %w", err)
	}
	if ltv.LTE(math.LegacyZeroDec()) || ltv.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("max_loan_to_value must be between 0 and 1")
	}

	return sdk.ValidateDenom(p.SupportedCollateralDenom)
}

func (p Params) GetLiquidationRatioAsDec() (math.LegacyDec, error) {
	return math.LegacyNewDecFromStr(p.LiquidationRatio)
}

func (p Params) GetLiquidationPenaltyAsDec() (math.LegacyDec, error) {
	return math.LegacyNewDecFromStr(p.LiquidationPenalty)
}

func (p Params) GetMaxLoanToValueAsDec() (math.LegacyDec, error) {
	return math.LegacyNewDecFromStr(p.MaxLoanToValue)
}

func (p Params) GetMinDebtAmountAsInt() (math.Int, error) {
	amount, ok := math.NewIntFromString(p.MinDebtAmount)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid min debt amount: %s", p.MinDebtAmount)
	}
	return amount, nil
}
