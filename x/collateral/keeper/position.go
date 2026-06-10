package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// SetPosition stores a position
func (k Keeper) SetPosition(ctx sdk.Context, position types.Position) error {
	return k.Positions.Set(ctx, position.Owner, position)
}

// GetPosition retrieves a position by owner
func (k Keeper) GetPosition(ctx sdk.Context, owner string) (types.Position, error) {
	position, err := k.Positions.Get(ctx, owner)
	if err != nil {
		return types.Position{}, types.ErrPositionNotFound
	}
	return position, nil
}

// DeletePosition removes a position
func (k Keeper) DeletePosition(ctx sdk.Context, owner string) error {
	return k.Positions.Remove(ctx, owner)
}
