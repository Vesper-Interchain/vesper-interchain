package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// WithdrawCollateral returns uatom collateral from the module escrow account back
// to the owner's wallet. The oracle price must be fresh before any withdrawal is
// allowed so that the collateral ratio check uses current market data.
//
// Rules enforced:
//   - The withdrawn amount must not exceed the position's current collateral balance.
//   - If the position carries outstanding debt, the remaining collateral after
//     withdrawal must keep the collateral ratio at or above the liquidation ratio.
//   - If collateral drops to zero and debt is also zero, the position is deleted.
//
// The rewards module is notified of the share change in both the partial and
// full-close paths so reward accrual is always consistent with on-chain collateral.
func (k Keeper) WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Require a fresh oracle price before performing any ratio check.
	if err := k.CheckOraclePrice(ctx, params.SupportedCollateralDenom); err != nil {
		return err
	}

	position, err := k.GetPosition(ctx, owner.String())
	if err != nil {
		return err
	}

	currentCollateral, ok := math.NewIntFromString(position.CollateralAmount)
	if !ok {
		return types.ErrInvalidAmount
	}
	if amount.GT(currentCollateral) {
		return types.ErrInsufficientCollateral
	}

	newCollateral := currentCollateral.Sub(amount)

	currentDebt, ok := math.NewIntFromString(position.DebtAmount)
	if !ok {
		return types.ErrInvalidAmount
	}

	// Only run the health check when debt is outstanding. A position with zero
	// debt can have all collateral withdrawn regardless of the remaining balance.
	if currentDebt.IsPositive() {
		tempPosition := position
		tempPosition.CollateralAmount = newCollateral.String()
		isHealthy, err := k.IsPositionHealthy(ctx, tempPosition)
		if err != nil {
			return err
		}
		if !isHealthy {
			return types.ErrCollateralRatioTooLow
		}
	}

	// Release the requested collateral from module escrow to the owner.
	coins := sdk.NewCoins(sdk.NewCoin(params.SupportedCollateralDenom, amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, coins); err != nil {
		return err
	}

	if newCollateral.IsZero() && currentDebt.IsZero() {
		// Full close: no remaining collateral and no outstanding debt — remove position.
		if err := k.DeletePosition(ctx, owner.String()); err != nil {
			return err
		}
		// Signal zero shares to the rewards module so the user stops accruing.
		k.notifyRewards(ctx, owner, math.ZeroInt())
	} else {
		// Partial withdrawal: update the stored collateral balance.
		position.CollateralAmount = newCollateral.String()
		position.UpdatedAt = ctx.BlockTime().Unix()
		if err := k.SetPosition(ctx, position); err != nil {
			return err
		}
		k.notifyRewards(ctx, owner, newCollateral)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCollateral,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}
