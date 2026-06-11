package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// SetPosition writes a Position to the Positions collection, keyed by the
// owner's bech32 address string. This is the single write path for all vault
// state changes (deposit, withdraw, mint, repay, liquidate).
func (k Keeper) SetPosition(ctx sdk.Context, position types.Position) error {
	return k.Positions.Set(ctx, position.Owner, position)
}

// GetPosition retrieves the Position for a given owner address string.
// Returns ErrPositionNotFound if the owner has never opened a vault or if the
// vault was previously deleted (fully repaid and withdrawn).
func (k Keeper) GetPosition(ctx sdk.Context, owner string) (types.Position, error) {
	position, err := k.Positions.Get(ctx, owner)
	if err != nil {
		return types.Position{}, types.ErrPositionNotFound
	}
	return position, nil
}

// DeletePosition removes a Position from state. Called when a vault is fully
// closed (zero collateral and zero debt) to reclaim storage.
func (k Keeper) DeletePosition(ctx sdk.Context, owner string) error {
	return k.Positions.Remove(ctx, owner)
}
