package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

/*
@fix: GenesisState now requires TotalMinted and TotalBurned fields (string-encoded math.Int)
      alongside Params. These track cumulative mint/burn volume for observability. An empty
      GenesisState{} would fail Validate() because TotalMinted/TotalBurned would be blank
      strings that cannot be parsed as math.Int.
@fix: InitGenesis and ExportGenesis signatures changed to sdk.Context (not context.Context).
      sdk.UnwrapSDKContext converts the test's context.Context back to sdk.Context.
*/
func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:      types.DefaultParams(),
		TotalMinted: "0",
		TotalBurned: "0",
	}

	f := initFixture(t)
	// @fix: unwrap required — InitGenesis/ExportGenesis use sdk.Context not context.Context
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	f.keeper.InitGenesis(sdkCtx, genesisState)
	got := f.keeper.ExportGenesis(sdkCtx)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
}
