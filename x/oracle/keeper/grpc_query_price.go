package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// Price handles the gRPC query for a single denomination's current price.
// Returns codes.NotFound if no price has been submitted for the requested denom.
func (q queryServer) Price(ctx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	price, err := q.k.GetPrice(sdkCtx, req.Denom)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "price not found for denom: %s", req.Denom)
	}

	return &types.QueryPriceResponse{Price: &price}, nil
}

// Prices handles the gRPC query for all currently stored prices.
// Converts the internal []Price slice to []*Price because the protobuf response
// type requires pointer elements.
func (q queryServer) Prices(ctx context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	prices, err := q.k.GetAllPrices(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Protobuf repeated fields use pointer semantics; convert value slice to pointer slice.
	pricePtrs := make([]*types.Price, len(prices))
	for i := range prices {
		pricePtrs[i] = &prices[i]
	}

	return &types.QueryPricesResponse{Prices: pricePtrs}, nil
}
