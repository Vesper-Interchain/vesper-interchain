package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MintStablecoin creates new uvusd tokens and credits them to the recipient.
// This function is called exclusively by the collateral keeper after it has
// verified that the user has sufficient collateral to back the new debt.
//
// Steps:
//  1. Load the stablecoin denom from Params.
//  2. Validate amount and recipient.
//  3. Mint coins into the module account via the bank keeper.
//  4. Transfer the minted coins from the module account to the recipient.
//  5. Increment the lifetime TotalMinted counter.
//  6. Emit a mint event.
func (k Keeper) MintStablecoin(ctx context.Context, recipient sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	if params.StableDenom == "" {
		return fmt.Errorf("invalid stable denom")
	}
	if !amount.IsPositive() {
		return types.ErrInvalidAmount
	}
	if recipient.Empty() {
		return types.ErrInvalidAmount
	}

	coin := sdk.NewCoin(params.StableDenom, amount)
	if coin.Denom != params.StableDenom {
		return types.ErrInvalidDenom
	}
	coins := sdk.NewCoins(coin)

	// Create tokens in the module account; these do not yet exist on any balance.
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Move the freshly minted tokens from the module account to the recipient.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
		return err
	}

	// Keep the lifetime counter in sync with every mint.
	if err := k.updateMinted(ctx, amount); err != nil {
		return err
	}

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
