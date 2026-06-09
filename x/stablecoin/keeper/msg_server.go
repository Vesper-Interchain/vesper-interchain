package keeper

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

type msgServer struct {
	Keeper Keeper
}

func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = (*msgServer)(nil)

func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert authority string to bytes for comparison
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
