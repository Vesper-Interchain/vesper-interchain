package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// MintStablecoin issues new uvusd tokens to the vault owner against their
// locked collateral. The oracle price must be fresh before minting because the
// mintable ceiling is derived from the current collateral-to-USD value.
//
// Rules enforced:
//   - The oracle price must not be stale.
//   - The amount must be positive.
//   - The resulting debt must not exceed MaxLTV * collateral_value_usd.
//   - The resulting debt must meet the minimum debt floor (Params.MinDebtAmount)
//     to deter positions that are too small to be profitable to liquidate.
//
// On success the stablecoin keeper mints new uvusd directly into the owner's wallet.
func (k Keeper) MintStablecoin(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Require a fresh oracle price so the LTV calculation uses current market data.
	if err := k.CheckOraclePrice(ctx, params.SupportedCollateralDenom); err != nil {
		return err
	}

	if !amount.IsPositive() {
		return types.ErrInvalidAmount
	}

	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	// GetMintableAmount returns (maxDebt - currentDebt), i.e. how much more
	// uvusd the user is still allowed to borrow given their collateral.
	mintable, err := k.GetMintableAmount(ctx, position)
	if err != nil {
		return err
	}
	if amount.GT(mintable) {
		return fmt.Errorf("%w: max %s", types.ErrMintLimitExceeded, mintable.String())
	}

	currentDebt, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return types.ErrInvalidAmount
	}
	newDebt := currentDebt.Add(amount)

	// Enforce the minimum debt floor to keep positions economically viable for liquidation.
	minDebt, err := params.GetMinDebtAmountAsInt()
	if err != nil {
		return err
	}
	if newDebt.LT(minDebt) && !newDebt.IsZero() {
		return types.ErrDebtTooLow
	}

	// Persist the updated debt before calling the stablecoin keeper so that if
	// the mint fails the position state is not permanently dirty.
	position.DebtAmount = newDebt.String()
	position.UpdatedAt = ctx.BlockTime().Unix()
	if err := k.SetPosition(ctx, position); err != nil {
		return err
	}

	// Delegate the actual token minting to the stablecoin keeper.
	if err := k.stablecoinKeeper.MintStablecoin(ctx, owner, amount); err != nil {
		return err
	}

	collateralRatio, _ := k.GetCollateralRatio(ctx, position)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintStablecoin,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDebt, position.DebtAmount),
			sdk.NewAttribute(types.AttributeKeyCollateralRatio, collateralRatio.String()),
		),
	)

	return nil
}
