package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

func (k msgServer) MintStablecoin(ctx context.Context, msg *types.MsgMintStablecoin) (*types.MsgMintStablecoinResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgMintStablecoinResponse{}, nil
}
