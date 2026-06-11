package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BurnStablecoin destroys uvusd from the owner's wallet. It is called by the
// collateral keeper during debt repayment and liquidation.
//
// Steps:
//  1. Validate the denom, amount, and owner.
//  2. Perform an underflow guard: TotalBurned + amount must not exceed TotalMinted.
//  3. Send tokens from the owner's account to the module account.
//  4. Burn the tokens from the module account.
//  5. Increment the lifetime TotalBurned counter.
//  6. Emit a burn event.
//
// Important: this function handles the send-to-module AND burn internally.
// Callers must NOT issue a separate SendCoinsFromAccountToModule before calling
// BurnStablecoin; doing so would cause a double-debit of the owner's balance.
func (k Keeper) BurnStablecoin(ctx context.Context, owner sdk.AccAddress, amount math.Int) error {
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
	if owner.Empty() {
		return types.ErrInvalidAmount
	}

	coin := sdk.NewCoin(params.StableDenom, amount)
	if coin.Denom != params.StableDenom {
		return types.ErrInvalidDenom
	}
	coins := sdk.NewCoins(coin)

	// Guard against burning more than was ever minted, which would indicate a
	// serious accounting bug and must be caught before any state mutation.
	currentMinted, err := k.GetTotalMinted(ctx)
	if err != nil {
		return err
	}
	currentBurned, err := k.GetTotalBurned(ctx)
	if err != nil {
		return err
	}
	if currentBurned.Add(amount).GT(currentMinted) {
		return types.ErrSupplyUnderflow
	}

	// Move tokens from the owner to the module account before burning.
	// The bank module requires coins to be in a module account before BurnCoins.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	// Destroy the tokens; they are removed from the total supply permanently.
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Keep the lifetime burned counter in sync with every burn.
	if err := k.updateBurned(ctx, amount); err != nil {
		return err
	}

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
