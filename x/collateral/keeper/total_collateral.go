package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// GetParams is a convenience wrapper that exposes module parameters through the
// context.Context interface, allowing callers that hold a context.Context (rather
// than sdk.Context) to read params without an extra type assertion.
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.Params.Get(ctx)
}

// GetTotalCollateral iterates all open positions and returns the sum of their
// CollateralAmount fields (in uatom). This value represents the total assets
// under management by the collateral module at a given block height.
// Used by the rewards module to compute per-block share fractions.
func (k Keeper) GetTotalCollateral(ctx sdk.Context) (math.Int, error) {
	total := math.ZeroInt()

	iter, err := k.Positions.Iterate(ctx, nil)
	if err != nil {
		return math.ZeroInt(), err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return math.ZeroInt(), err
		}
		amount, ok := math.NewIntFromString(kv.Value.CollateralAmount)
		if !ok {
			// Skip positions with corrupted collateral strings rather than aborting.
			continue
		}
		total = total.Add(amount)
	}

	return total, nil
}
