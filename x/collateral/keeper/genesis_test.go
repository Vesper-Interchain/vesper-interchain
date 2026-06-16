package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

/*
@fix: InitGenesis and ExportGenesis were updated to accept sdk.Context instead of
      context.Context. The test previously called them with f.ctx (a context.Context)
      and checked error returns; the actual keeper signatures return no error and require
      sdk.Context, so we unwrap with sdk.UnwrapSDKContext and drop the error checks.
*/
func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	f := initFixture(t)
	// @fix: UnwrapSDKContext required — keeper methods use sdk.Context not context.Context
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	f.keeper.InitGenesis(sdkCtx, genesisState)
	got := f.keeper.ExportGenesis(sdkCtx)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
}
