package types_test

import (
	"testing"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
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
			// @fix: Empty GenesisState{} is no longer valid — Validate() now checks that
			// TotalMinted and TotalBurned are parseable math.Int strings and that
			// Params.StableDenom is non-empty. Must supply DefaultParams() and zero strings
			// to form a valid genesis that isn't the exact DefaultGenesis() case.
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params:      types.DefaultParams(),
				TotalMinted: "0",
				TotalBurned: "0",
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
