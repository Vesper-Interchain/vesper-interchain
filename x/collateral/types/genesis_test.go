package types_test

import (
	"testing"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			// @fix: Empty GenesisState{} is no longer valid — Params.Validate() now rejects
			// a zero-value Params struct because liquidation_ratio would be unset.
			// Must supply DefaultParams() to produce a valid non-default genesis entry.
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
			},
			valid: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
