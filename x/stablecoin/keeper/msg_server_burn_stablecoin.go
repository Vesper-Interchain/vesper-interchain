package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

func (k msgServer) BurnStablecoin(ctx context.Context, msg *types.MsgBurnStablecoin) (*types.MsgBurnStablecoinResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgBurnStablecoinResponse{}, nil
}
