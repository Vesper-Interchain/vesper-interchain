package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/rewards/types"
)

// InitGenesis initialises the rewards module state from the provided genesis state.
// The accumulator and all per-user shares start at zero; the collateral module will
// call UpdateShares as users deposit, rebuilding the share distribution naturally.
func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) {
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis serialises the rewards module state into a GenesisState struct
// for snapshotting or chain migration. Only Params are exported because the
// accumulator and per-user shares are derived from live chain state and will
// be rebuilt on import via InitGenesis + collateral deposit callbacks.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	params, _ := k.GetParams(ctx)
	return types.GenesisState{Params: params}
}
