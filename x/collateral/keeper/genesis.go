package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := data.Params.Validate(); err != nil {
		panic(err)
	}
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		Params: params,
	}
}
