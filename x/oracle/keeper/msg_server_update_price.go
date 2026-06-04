package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

func (k msgServer) UpdatePrice(ctx context.Context, msg *types.MsgUpdatePrice) (*types.MsgUpdatePriceResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgUpdatePriceResponse{}, nil
}
