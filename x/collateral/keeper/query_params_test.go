package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/keeper"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

/*
@fix: Renamed NewQueryServerImpl → NewQueryServer to match the updated keeper export name.
@fix: QueryParamsResponse.Params changed from value type (types.Params) to pointer (*types.Params)
      so that protobuf optional semantics are preserved (a nil pointer means "not set" vs.
      a zero-value struct which is ambiguous). The assertion now uses &params.
*/
func TestParamsQuery(t *testing.T) {
	f := initFixture(t)

	qs := keeper.NewQueryServer(f.keeper)
	params := types.DefaultParams()
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	response, err := qs.Params(f.ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: &params}, response)
}
