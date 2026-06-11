package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// UpdatePrice processes a MsgUpdatePrice transaction. It enforces that only the
// single designated oracle address stored in module Params may submit price data.
// This design intentionally avoids a multi-validator oracle aggregation scheme to
// keep the implementation simple for a portfolio chain; production deployments
// would replace this with a median-aggregated feed from multiple providers.
func (m msgServer) UpdatePrice(goCtx context.Context, msg *types.MsgUpdatePrice) (*types.MsgUpdatePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := m.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Reject the message if the signer is not the authorised oracle feeder.
	if msg.Creator != params.OracleAddress {
		return nil, types.ErrUnauthorizedOracle
	}

	// Stamp the price with the current block time so consumers can detect staleness
	// using a deterministic, consensus-driven clock rather than a wall-clock.
	price := types.Price{
		Denom:     msg.Denom,
		Price:     msg.Price,
		Timestamp: ctx.BlockTime().Unix(),
		Source:    msg.Source,
	}

	if err = m.SetPrice(ctx, price); err != nil {
		return nil, err
	}

	// Emit a structured event so off-chain indexers and monitoring services can
	// track price feed activity without scanning raw transaction bytes.
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
