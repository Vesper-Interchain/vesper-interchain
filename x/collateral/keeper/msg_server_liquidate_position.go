package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k msgServer) LiquidatePosition(ctx context.Context, msg *types.MsgLiquidatePosition) (*types.MsgLiquidatePositionResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgLiquidatePositionResponse{}, nil
}
