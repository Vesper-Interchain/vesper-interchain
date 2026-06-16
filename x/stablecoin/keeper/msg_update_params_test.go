package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/keeper"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

/*
@fix: Renamed NewMsgServerImpl → NewMsgServer to match the updated keeper export.
@fix: Empty Params{} now expects an error "stable_denom cannot be empty" because
      Params.Validate() was tightened — an empty stable_denom means MintStablecoin
      would create coins with a blank denom, which the bank module rejects at runtime.
      Better to catch this at param update time than at mint time.
*/
func TestMsgUpdateParams(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServer(f.keeper)

	params := types.DefaultParams()
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	authorityStr, err := f.addressCodec.BytesToString(f.keeper.GetAuthority())
	require.NoError(t, err)

	// default params
	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid params - empty denom",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params:    types.Params{},
			},
			expErr:    true,
			expErrMsg: "stable_denom cannot be empty",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: authorityStr,
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.UpdateParams(f.ctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
