package keeper

import (
	"cosmossdk.io/math"
	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPrice validates and persists a price submission for the given denomination.
// The price must be a non-negative decimal string and the source must be non-empty.
// The timestamp is set externally (by the message handler) to the block time so that
// staleness checks remain deterministic across all nodes.
func (k Keeper) SetPrice(ctx sdk.Context, price types.Price) error {
	if price.Denom == "" {
		return types.ErrEmptyDenom
	}

	priceDec, err := math.LegacyNewDecFromStr(price.Price)
	if err != nil {
		return types.ErrInvalidPrice.Wrapf("invalid price string: %s", price.Price)
	}
	if priceDec.IsNegative() {
		return types.ErrInvalidPrice.Wrap("price cannot be negative")
	}

	if price.Source == "" {
		return types.ErrEmptySource
	}

	return k.Prices.Set(ctx, price.Denom, price)
}

// GetPrice retrieves the most recently stored price for a denomination.
// Returns ErrPriceNotFound if no price has been submitted for that denom yet.
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

// RemovePrice deletes the stored price for a denomination. Used during genesis
// migration or administrative cleanup; normal operation does not remove prices.
func (k Keeper) RemovePrice(ctx sdk.Context, denom string) error {
	if denom == "" {
		return types.ErrEmptyDenom
	}
	return k.Prices.Remove(ctx, denom)
}

// GetAllPrices returns every price currently stored in state. Used by the query
// server to implement the Prices RPC and by genesis export.
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

// HasPrice reports whether a price entry exists for the given denomination.
// Used as a lightweight existence check before calling GetPrice.
func (k Keeper) HasPrice(ctx sdk.Context, denom string) bool {
	_, err := k.Prices.Get(ctx, denom)
	return err == nil
}

// GetPriceValue returns the stored price for a denomination as a LegacyDec.
// Other modules (x/collateral) call this to obtain the numeric value for
// collateral ratio calculations.
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

// IsPriceStale reports whether the stored price for a denomination is older than
// maxAgeSeconds relative to the current block time. The collateral module calls
// this before every vault operation to prevent stale-price exploitation.
func (k Keeper) IsPriceStale(ctx sdk.Context, denom string, maxAgeSeconds int64) (bool, error) {
	price, err := k.GetPrice(ctx, denom)
	if err != nil {
		// If no price exists at all, treat it as stale to block operations.
		return true, err
	}

	ageSeconds := ctx.BlockTime().Unix() - price.Timestamp
	return ageSeconds > maxAgeSeconds, nil
}
