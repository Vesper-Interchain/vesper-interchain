package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/keeper"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

/*
@fix: Renamed NewQueryServerImpl → NewQueryServer to match the updated keeper export.
@fix: QueryParamsResponse.Params field changed from value type (types.Params) to pointer
      (*types.Params) to match the updated proto-generated struct. A pointer field lets
      gRPC signal "params not set" (nil) vs. "params are zero-valued" (non-nil, zero).
      The assertion is updated to compare &params (pointer to local variable).
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
