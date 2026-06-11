// Package keeper implements the stablecoin module's state management.
// The stablecoin module is a thin mint-and-burn layer; it does not decide
// *when* to mint or burn — that is the collateral module's responsibility.
// This module only executes the token operations and tracks the running
// supply counters (TotalMinted, TotalBurned) for protocol-level accounting.
package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

// Keeper is the stablecoin module keeper. It holds the module's KV store
// collections and exposes the mint/burn interface consumed by x/collateral.
type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec

	// authority is the governance address (x/gov module account) that can
	// update module parameters such as the stablecoin denom.
	authority []byte

	// Schema is the collections schema, needed for genesis import/export.
	Schema collections.Schema

	// Params holds governance-controlled configuration (currently just StableDenom).
	Params collections.Item[types.Params]

	// Supply tracks lifetime TotalMinted and TotalBurned values for the module.
	// Circulating supply = TotalMinted - TotalBurned.
	Supply collections.Item[types.SupplyState]

	bankKeeper types.BankKeeper
}

// NewKeeper constructs a new stablecoin Keeper and registers all collections.
func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bankKeeper types.BankKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bankKeeper:   bankKeeper,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Supply:       collections.NewItem(sb, types.SupplyKey, "supply", codec.CollValue[types.SupplyState](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the authority address that can update module parameters.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// GetSupply returns the current SupplyState. If no supply record exists yet
// (fresh chain, no mints performed), it returns zero values for both counters
// rather than an error.
func (k Keeper) GetSupply(ctx context.Context) (types.SupplyState, error) {
	supply, err := k.Supply.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.SupplyState{TotalMinted: "0", TotalBurned: "0"}, nil
		}
		return types.SupplyState{}, err
	}
	return supply, nil
}

// GetTotalMinted returns the cumulative amount of uvusd ever minted as math.Int.
func (k Keeper) GetTotalMinted(ctx context.Context) (math.Int, error) {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	minted, ok := math.NewIntFromString(supply.TotalMinted)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid total minted: %s", supply.TotalMinted)
	}
	return minted, nil
}

// GetTotalBurned returns the cumulative amount of uvusd ever burned as math.Int.
func (k Keeper) GetTotalBurned(ctx context.Context) (math.Int, error) {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	burned, ok := math.NewIntFromString(supply.TotalBurned)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid total burned: %s", supply.TotalBurned)
	}
	return burned, nil
}

// GetCirculatingSupply returns the net outstanding uvusd supply (minted - burned).
func (k Keeper) GetCirculatingSupply(ctx context.Context) (math.Int, error) {
	minted, err := k.GetTotalMinted(ctx)
	if err != nil {
		return math.Int{}, err
	}
	burned, err := k.GetTotalBurned(ctx)
	if err != nil {
		return math.Int{}, err
	}
	return minted.Sub(burned), nil
}

// updateMinted adds amount to the lifetime minted counter.
// Called after a successful bank.MintCoins so the supply accounting stays in sync.
func (k Keeper) updateMinted(ctx context.Context, amount math.Int) error {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return err
	}
	currentMinted, ok := math.NewIntFromString(supply.TotalMinted)
	if !ok {
		return fmt.Errorf("invalid total minted: %s", supply.TotalMinted)
	}
	supply.TotalMinted = currentMinted.Add(amount).String()
	return k.Supply.Set(ctx, supply)
}

// updateBurned adds amount to the lifetime burned counter.
// Called after a successful bank.BurnCoins so the supply accounting stays in sync.
func (k Keeper) updateBurned(ctx context.Context, amount math.Int) error {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return err
	}
	currentBurned, ok := math.NewIntFromString(supply.TotalBurned)
	if !ok {
		return fmt.Errorf("invalid total burned: %s", supply.TotalBurned)
	}
	supply.TotalBurned = currentBurned.Add(amount).String()
	return k.Supply.Set(ctx, supply)
}
