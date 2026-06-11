package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// GetCollateralValueUSD converts a raw collateral amount (in uatom) to its
// current USD value using the oracle price.
//
// Formula: value_usd = (amount_uatom / 1_000_000) * price_per_atom
//
// The 1_000_000 divisor converts micro-units to human-readable ATOM before
// multiplying by the price, which the oracle reports in USD per 1 ATOM.
func (k Keeper) GetCollateralValueUSD(ctx sdk.Context, amountStr string, denom string) (math.LegacyDec, error) {
	amount, ok := math.NewIntFromString(amountStr)
	if !ok {
		return math.LegacyDec{}, fmt.Errorf("invalid amount: %s", amountStr)
	}

	price, err := k.oracleKeeper.GetPriceValue(ctx, denom)
	if err != nil {
		return math.LegacyDec{}, err
	}

	// Convert uatom to ATOM (6 decimal places) before applying the price.
	amountHumanDec := math.LegacyNewDecFromInt(amount).QuoInt64(1_000_000)
	return amountHumanDec.Mul(price), nil
}

// GetCollateralRatio computes the current collateral ratio for a position.
//
// Formula: CR = collateral_value_usd / debt_value_usd
//
// Debt is stored as uvusd (1 uvusd = $0.000001) so it is divided by 1_000_000
// to obtain the dollar-denominated debt value.
//
// Returns a very large sentinel value (999999) when debt is zero to indicate
// an infinitely healthy position without causing a division-by-zero error.
func (k Keeper) GetCollateralRatio(ctx sdk.Context, position types.Position) (math.LegacyDec, error) {
	debt, err := math.LegacyNewDecFromStr(position.DebtAmount)
	if err != nil {
		return math.LegacyDec{}, err
	}
	debtUSD := debt.QuoInt64(1_000_000)

	if debtUSD.IsZero() {
		// No debt means an effectively infinite collateral ratio.
		return math.LegacyNewDecFromInt(math.NewInt(999999)), nil
	}

	collateralValueUSD, err := k.GetCollateralValueUSD(ctx, position.CollateralAmount, position.CollateralDenom)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return collateralValueUSD.Quo(debtUSD), nil
}

// IsPositionHealthy returns true when the position's collateral ratio is at or
// above the liquidation threshold defined in Params.LiquidationRatio.
// A ratio below this threshold allows any caller to liquidate the position.
func (k Keeper) IsPositionHealthy(ctx sdk.Context, position types.Position) (bool, error) {
	collateralRatio, err := k.GetCollateralRatio(ctx, position)
	if err != nil {
		return false, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return false, err
	}

	liquidationRatio, err := params.GetLiquidationRatioAsDec()
	if err != nil {
		return false, err
	}

	return collateralRatio.GTE(liquidationRatio), nil
}

// GetMintableAmount returns how many additional uvusd tokens the owner of a
// position can still borrow given their current collateral and existing debt.
//
// Formula: mintable = (collateral_value_usd * MaxLTV * 1_000_000) - current_debt_uvusd
//
// The result is always non-negative; if the user is already at or beyond the
// maximum debt ceiling, zero is returned.
func (k Keeper) GetMintableAmount(ctx sdk.Context, position types.Position) (math.Int, error) {
	collateralValueUSD, err := k.GetCollateralValueUSD(ctx, position.CollateralAmount, position.CollateralDenom)
	if err != nil {
		return math.Int{}, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.Int{}, err
	}

	maxLTV, err := params.GetMaxLoanToValueAsDec()
	if err != nil {
		return math.Int{}, err
	}

	// Maximum allowable debt in USD, then converted back to uvusd (×1_000_000).
	maxAllowedDebtUSD := collateralValueUSD.Mul(maxLTV)
	maxAllowedDebtUVUSD := maxAllowedDebtUSD.MulInt64(1_000_000)

	currentDebtUVUSD, err := math.LegacyNewDecFromStr(position.DebtAmount)
	if err != nil {
		return math.Int{}, err
	}

	available := maxAllowedDebtUVUSD.Sub(currentDebtUVUSD)
	if available.LTE(math.LegacyZeroDec()) {
		return math.ZeroInt(), nil
	}

	return available.TruncateInt(), nil
}

// CheckOraclePrice verifies that the stored oracle price for a denom is within
// the acceptable staleness window defined by Params.OraclePriceStaleSeconds.
// All state-changing vault operations call this as a guard to prevent the
// protocol from operating on outdated collateral valuations.
func (k Keeper) CheckOraclePrice(ctx sdk.Context, denom string) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	stale, err := k.oracleKeeper.IsPriceStale(ctx, denom, params.OraclePriceStaleSeconds)
	if err != nil {
		return err
	}
	if stale {
		return fmt.Errorf("oracle price for %s is stale", denom)
	}
	return nil
}

// CalculateLiquidationOutput computes what the liquidator receives and what debt
// they must repay in order to close an under-collateralised position.
//
// Formula:
//   - debtToRepay   = full outstanding uvusd debt
//   - totalValueUSD = debtUSD * (1 + liquidationPenalty)
//   - collateralOut = totalValueUSD / price_per_uatom
//
// If the position does not hold enough collateral to cover the full penalty
// amount, all remaining collateral is returned (partial collateral seizure).
func (k Keeper) CalculateLiquidationOutput(ctx sdk.Context, position types.Position) (collateralToGive math.Int, penaltyUSD math.LegacyDec, debtToRepayUVUSD math.Int, err error) {
	debtUVUSD, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return math.Int{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("invalid debt amount")
	}
	debtToRepayUVUSD = debtUVUSD

	debtUSD := math.LegacyNewDecFromInt(debtUVUSD).QuoInt64(1_000_000)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.Int{}, math.LegacyDec{}, math.Int{}, err
	}
	penaltyRate, err := params.GetLiquidationPenaltyAsDec()
	if err != nil {
		return math.Int{}, math.LegacyDec{}, math.Int{}, err
	}

	// Total value the liquidator is entitled to: debt + penalty bonus.
	totalValueUSD := debtUSD.Mul(math.LegacyOneDec().Add(penaltyRate))
	penaltyUSD = totalValueUSD.Sub(debtUSD)

	// Derive the equivalent collateral amount at the current oracle price.
	// Price is reported as USD per ATOM; collateral is in uatom.
	price, err := k.oracleKeeper.GetPriceValue(ctx, position.CollateralDenom)
	if err != nil {
		return math.Int{}, math.LegacyDec{}, math.Int{}, err
	}

	// collateralNeeded (uatom) = totalValueUSD * 1e6 * 1e6 / price
	// The first 1e6 converts USD to uvusd scale; the second converts ATOM to uatom.
	collateralNeededDec := totalValueUSD.MulInt64(1_000_000).MulInt64(1_000_000).Quo(price)
	collateralNeeded := collateralNeededDec.TruncateInt()

	availableCollateral, ok := math.NewIntFromString(position.CollateralAmount)
	if !ok {
		return math.Int{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("invalid collateral amount")
	}

	// Cap at available collateral to handle positions where the penalty would
	// require more collateral than actually exists in the vault.
	if availableCollateral.LT(collateralNeeded) {
		return availableCollateral, penaltyUSD, debtToRepayUVUSD, nil
	}

	return collateralNeeded, penaltyUSD, debtToRepayUVUSD, nil
}
