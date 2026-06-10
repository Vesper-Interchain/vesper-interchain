package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k Keeper) MintStablecoin(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
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

	minDebt, err := params.GetMinDebtAmountAsInt()
	if err != nil {
		return err
	}
	if newDebt.LT(minDebt) && !newDebt.IsZero() {
		return types.ErrDebtTooLow
	}

	position.DebtAmount = newDebt.String()
	position.UpdatedAt = ctx.BlockTime().Unix()
	if err := k.SetPosition(ctx, position); err != nil {
		return err
	}

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
