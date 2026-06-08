package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// UpdatePrice handles the UpdatePrice message
func (m msgServer) UpdatePrice(goCtx context.Context, msg *types.MsgUpdatePrice) (*types.MsgUpdatePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := m.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Creator != params.OracleAddress {
		return nil, types.ErrUnauthorizedOracle
	}

	price := types.Price{
		Denom:     msg.Denom,
		Price:     msg.Price,
		Timestamp: ctx.BlockTime().Unix(),
		Source:    msg.Source,
	}

	err = m.SetPrice(ctx, price)
	if err != nil {
		return nil, err
	}

	// Emit event for off-chain tracking
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePriceUpdate,
			sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
			sdk.NewAttribute(types.AttributeKeyPrice, msg.Price),
			sdk.NewAttribute(types.AttributeKeySource, msg.Source),
			sdk.NewAttribute(types.AttributeKeyTimestamp, strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
		),
	)

	return &types.MsgUpdatePriceResponse{}, nil
}
