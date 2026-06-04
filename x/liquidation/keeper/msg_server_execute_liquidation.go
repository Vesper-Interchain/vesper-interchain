package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/types"
)

func (k msgServer) ExecuteLiquidation(ctx context.Context, msg *types.MsgExecuteLiquidation) (*types.MsgExecuteLiquidationResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgExecuteLiquidationResponse{}, nil
}
