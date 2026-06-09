package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

// MintStablecoin creates new stablecoin tokens and sends them to the recipient
func (k Keeper) MintStablecoin(ctx context.Context, recipient sdk.AccAddress, amount math.Int) error {
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

	// Validate recipient
	if recipient.Empty() {
		return types.ErrInvalidAmount
	}

	// Create coin with denom safety check
	coin := sdk.NewCoin(params.StableDenom, amount)
	if coin.Denom != params.StableDenom {
		return types.ErrInvalidDenom
	}

	coins := sdk.NewCoins(coin)

	// Mint coins to module account
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Send coins from module to recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
		return err
	}

	// Update total minted supply
	if err := k.updateMinted(ctx, amount); err != nil {
		return err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}
