package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

// InitGenesis initializes the stablecoin module's genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	// Validate params
	if err := data.Params.Validate(); err != nil {
		panic(err)
	}

	// Set params
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	// Parse total minted
	minted := math.ZeroInt()
	if data.TotalMinted != "" {
		var ok bool
		minted, ok = math.NewIntFromString(data.TotalMinted)
		if !ok {
			panic("invalid total_minted value")
		}
	}

	// Parse total burned
	burned := math.ZeroInt()
	if data.TotalBurned != "" {
		var ok bool
		burned, ok = math.NewIntFromString(data.TotalBurned)
		if !ok {
			panic("invalid total_burned value")
		}
	}

	// Validate invariant: burned cannot exceed minted
	if burned.GT(minted) {
		panic("total_burned cannot exceed total_minted")
	}

	// Set supply state
	supply := types.SupplyState{
		TotalMinted: minted.String(),
		TotalBurned: burned.String(),
	}
	if err := k.Supply.Set(ctx, supply); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the stablecoin module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	supply, err := k.Supply.Get(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:      params,
		TotalMinted: supply.TotalMinted,
		TotalBurned: supply.TotalBurned,
	}
}
