package keeper

import (
	"bytes"
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

type msgServer struct {
	Keeper Keeper
}

func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = (*msgServer)(nil)

func (m msgServer) DepositCollateral(goCtx context.Context, msg *types.MsgDepositCollateral) (*types.MsgDepositCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amount, ok := math.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.DepositCollateral(ctx, creator, amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositCollateralResponse{
		Owner:           msg.Creator,
		CollateralAmount: msg.Amount,
	}, nil
}

func (m msgServer) WithdrawCollateral(goCtx context.Context, msg *types.MsgWithdrawCollateral) (*types.MsgWithdrawCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amount, ok := math.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.WithdrawCollateral(ctx, creator, amount); err != nil {
		return nil, err
	}

	position, err := m.Keeper.GetPosition(ctx, msg.Creator)
	if err != nil {
		return &types.MsgWithdrawCollateralResponse{WithdrawnAmount: msg.Amount}, nil
	}

	collateralRatio, _ := m.Keeper.GetCollateralRatio(ctx, position)

	return &types.MsgWithdrawCollateralResponse{
		WithdrawnAmount: msg.Amount,
		CollateralRatio: collateralRatio.String(),
	}, nil
}

func (m msgServer) MintStablecoin(goCtx context.Context, msg *types.MsgMintStablecoin) (*types.MsgMintStablecoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amount, ok := math.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.MintStablecoin(ctx, creator, amount); err != nil {
		return nil, err
	}

	position, err := m.Keeper.GetPosition(ctx, msg.Creator)
	if err != nil {
		return &types.MsgMintStablecoinResponse{MintedAmount: msg.Amount}, nil
	}

	collateralRatio, _ := m.Keeper.GetCollateralRatio(ctx, position)

	return &types.MsgMintStablecoinResponse{
		MintedAmount:    msg.Amount,
		CollateralRatio: collateralRatio.String(),
	}, nil
}

func (m msgServer) RepayDebt(goCtx context.Context, msg *types.MsgRepayDebt) (*types.MsgRepayDebtResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amount, ok := math.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, types.ErrInvalidAmount
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.RepayDebt(ctx, creator, amount); err != nil {
		return nil, err
	}

	position, err := m.Keeper.GetPosition(ctx, msg.Creator)
	if err != nil {
		return &types.MsgRepayDebtResponse{RemainingDebt: "0"}, nil
	}

	collateralRatio, _ := m.Keeper.GetCollateralRatio(ctx, position)

	return &types.MsgRepayDebtResponse{
		RemainingDebt:   position.DebtAmount,
		CollateralRatio: collateralRatio.String(),
	}, nil
}

func (m msgServer) LiquidatePosition(goCtx context.Context, msg *types.MsgLiquidatePosition) (*types.MsgLiquidatePositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.LiquidatePosition(ctx, owner, liquidator); err != nil {
		return nil, err
	}

	return &types.MsgLiquidatePositionResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityBytes := m.Keeper.GetAuthority()
	reqAuthorityBytes, err := m.Keeper.addressCodec.StringToBytes(req.Authority)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", req.Authority)
	}

	if !bytes.Equal(authorityBytes, reqAuthorityBytes) {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", string(authorityBytes), req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	if err := m.Keeper.Params.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
