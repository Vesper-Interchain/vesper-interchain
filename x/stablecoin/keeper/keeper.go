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

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec

	authority []byte

	Schema collections.Schema

	Params collections.Item[types.Params]
	Supply collections.Item[types.SupplyState]

	bankKeeper types.BankKeeper
}

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

// GetAuthority returns the module's authority (x/gov module address)
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// GetSupply returns the current supply state
func (k Keeper) GetSupply(ctx context.Context) (types.SupplyState, error) {
	supply, err := k.Supply.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.SupplyState{
				TotalMinted: "0",
				TotalBurned: "0",
			}, nil
		}
		return types.SupplyState{}, err
	}
	return supply, nil
}

// GetTotalMinted returns total minted supply as math.Int
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

// GetTotalBurned returns total burned supply as math.Int
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

// GetCirculatingSupply returns minted - burned
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

// updateMinted adds to total minted supply
func (k Keeper) updateMinted(ctx context.Context, amount math.Int) error {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return err
	}

	currentMinted, ok := math.NewIntFromString(supply.TotalMinted)
	if !ok {
		return fmt.Errorf("invalid total minted: %s", supply.TotalMinted)
	}
	newMinted := currentMinted.Add(amount)
	supply.TotalMinted = newMinted.String()

	return k.Supply.Set(ctx, supply)
}

// updateBurned adds to total burned supply
func (k Keeper) updateBurned(ctx context.Context, amount math.Int) error {
	supply, err := k.GetSupply(ctx)
	if err != nil {
		return err
	}

	currentBurned, ok := math.NewIntFromString(supply.TotalBurned)
	if !ok {
		return fmt.Errorf("invalid total burned: %s", supply.TotalBurned)
	}
	newBurned := currentBurned.Add(amount)
	supply.TotalBurned = newBurned.String()

	return k.Supply.Set(ctx, supply)
}
