package keeper

import (
	"cosmossdk.io/math"
	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPrice stores a price submission for a denomination
func (k Keeper) SetPrice(ctx sdk.Context, price types.Price) error {
	// Validate denom
	if price.Denom == "" {
		return types.ErrEmptyDenom
	}

	// Validate price is a valid decimal
	priceDec, err := math.LegacyNewDecFromStr(price.Price)
	if err != nil {
		return types.ErrInvalidPrice.Wrapf("invalid price string: %s", price.Price)
	}

	if priceDec.IsNegative() {
		return types.ErrInvalidPrice.Wrap("price cannot be negative")
	}

	// Validate source
	if price.Source == "" {
		return types.ErrEmptySource
	}

	// Store the price using the Prices collection
	return k.Prices.Set(ctx, price.Denom, price)
}

// GetPrice retrieves the latest price for a denomination
func (k Keeper) GetPrice(ctx sdk.Context, denom string) (types.Price, error) {
	if denom == "" {
		return types.Price{}, types.ErrEmptyDenom
	}

	price, err := k.Prices.Get(ctx, denom)
	if err != nil {
		return types.Price{}, types.ErrPriceNotFound.Wrapf("denom: %s", denom)
	}
	return price, nil
}

// RemovePrice removes a price for a denomination
func (k Keeper) RemovePrice(ctx sdk.Context, denom string) error {
	if denom == "" {
		return types.ErrEmptyDenom
	}

	return k.Prices.Remove(ctx, denom)
}

// GetAllPrices returns all stored prices
func (k Keeper) GetAllPrices(ctx sdk.Context) ([]types.Price, error) {
	iter, err := k.Prices.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var prices []types.Price
	for ; iter.Valid(); iter.Next() {
		price, err := iter.Value()
		if err != nil {
			continue
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// HasPrice checks if a price exists for a denomination
func (k Keeper) HasPrice(ctx sdk.Context, denom string) bool {
	_, err := k.Prices.Get(ctx, denom)
	return err == nil
}

// GetPriceValue returns the price as a decimal for a denomination
func (k Keeper) GetPriceValue(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	price, err := k.GetPrice(ctx, denom)
	if err != nil {
		return math.LegacyDec{}, err
	}

	priceDec, err := math.LegacyNewDecFromStr(price.Price)
	if err != nil {
		return math.LegacyDec{}, types.ErrInvalidPrice.Wrapf("invalid price decimal for %s: %s", denom, price.Price)
	}

	return priceDec, nil
}

// IsPriceStale checks if a price is stale (older than maxAgeSeconds)
func (k Keeper) IsPriceStale(ctx sdk.Context, denom string, maxAgeSeconds int64) (bool, error) {
	price, err := k.GetPrice(ctx, denom)
	if err != nil {
		return true, err
	}

	currentTime := ctx.BlockTime().Unix()
	ageSeconds := currentTime - price.Timestamp

	return ageSeconds > maxAgeSeconds, nil
}
