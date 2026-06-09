package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

// BurnStablecoin destroys stablecoin tokens from the user's account
func (k Keeper) BurnStablecoin(ctx context.Context, owner sdk.AccAddress, amount math.Int) error {
	// Get params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Validate denom
	if params.StableDenom == "" {
		return fmt.Errorf("invalid stable denom")
	}

	// Validate amount
	if !amount.IsPositive() {
		return types.ErrInvalidAmount
	}

	// Validate owner
	if owner.Empty() {
		return types.ErrInvalidAmount
	}

	// Create coin with denom safety check
	coin := sdk.NewCoin(params.StableDenom, amount)
	if coin.Denom != params.StableDenom {
		return types.ErrInvalidDenom
	}

	coins := sdk.NewCoins(coin)

	// Check total minted underflow protection
	currentMinted, err := k.GetTotalMinted(ctx)
	if err != nil {
		return err
	}
	currentBurned, err := k.GetTotalBurned(ctx)
	if err != nil {
		return err
	}

	// Cannot burn more than minted
	if currentBurned.Add(amount).GT(currentMinted) {
		return types.ErrSupplyUnderflow
	}

	// Send coins from user to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	// Burn coins from module account
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Update total burned supply
	if err := k.updateBurned(ctx, amount); err != nil {
		return err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.AttributeKeySender, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}
