package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// GetCollateralValueUSD calculates the USD value of collateral
// Formula: value = (collateral_amount / 10^collateral_decimals) * price
func (k Keeper) GetCollateralValueUSD(ctx sdk.Context, amountStr string, denom string) (math.LegacyDec, error) {
	amount, ok := math.NewIntFromString(amountStr)
	if !ok {
		return math.LegacyDec{}, fmt.Errorf("invalid amount: %s", amountStr)
	}

	price, err := k.oracleKeeper.GetPriceValue(ctx, denom)
	if err != nil {
		return math.LegacyDec{}, err
	}

	// Convert amount to human readable (ATOM has 6 decimals)
	amountHumanDec := math.LegacyNewDecFromInt(amount).QuoInt64(1_000_000)

	return amountHumanDec.Mul(price), nil
}

// GetCollateralRatio calculates the collateral ratio of a position
// Collateral Ratio = Collateral Value (USD) / Debt (USD)
func (k Keeper) GetCollateralRatio(ctx sdk.Context, position types.Position) (math.LegacyDec, error) {
	debt, err := math.LegacyNewDecFromStr(position.DebtAmount)
	if err != nil {
		return math.LegacyDec{}, err
	}
	debtUSD := debt.QuoInt64(1_000_000)

	if debtUSD.IsZero() {
		return math.LegacyNewDecFromInt(math.NewInt(999999)), nil
	}

	collateralValueUSD, err := k.GetCollateralValueUSD(ctx, position.CollateralAmount, position.CollateralDenom)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return collateralValueUSD.Quo(debtUSD), nil
}

// IsPositionHealthy checks if collateral ratio >= liquidation ratio
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

// GetMintableAmount calculates how much uvUSD can be minted
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

// CheckOraclePrice checks if oracle price is stale
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

// CalculateLiquidationOutput calculates what liquidator gets
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

	totalValueUSD := debtUSD.Mul(math.LegacyOneDec().Add(penaltyRate))
	penaltyUSD = totalValueUSD.Sub(debtUSD)

	price, err := k.oracleKeeper.GetPriceValue(ctx, position.CollateralDenom)
	if err != nil {
		return math.Int{}, math.LegacyDec{}, math.Int{}, err
	}

	// Collateral needed = (Total USD * 1,000,000 * 1,000,000) / price
	collateralNeededDec := totalValueUSD.MulInt64(1_000_000).MulInt64(1_000_000).Quo(price)
	collateralNeeded := collateralNeededDec.TruncateInt()

	availableCollateral, ok := math.NewIntFromString(position.CollateralAmount)
	if !ok {
		return math.Int{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("invalid collateral amount")
	}

	if availableCollateral.LT(collateralNeeded) {
		return availableCollateral, penaltyUSD, debtToRepayUVUSD, nil
	}

	return collateralNeeded, penaltyUSD, debtToRepayUVUSD, nil
}
