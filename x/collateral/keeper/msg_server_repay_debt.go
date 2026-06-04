package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

func (k msgServer) RepayDebt(ctx context.Context, msg *types.MsgRepayDebt) (*types.MsgRepayDebtResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgRepayDebtResponse{}, nil
}
