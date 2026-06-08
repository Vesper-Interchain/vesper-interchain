package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// Price handles the query for a single price
func (q queryServer) Price(ctx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	price, err := q.k.GetPrice(sdkCtx, req.Denom)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "price not found for denom: %s", req.Denom)
	}

	// Price in response is *types.Price
	return &types.QueryPriceResponse{Price: &price}, nil
}

// Prices handles the query for all prices with pagination
func (q queryServer) Prices(ctx context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	prices, err := q.k.GetAllPrices(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert []Price to []*Price because protobuf expects pointers
	pricePtrs := make([]*types.Price, len(prices))
	for i := range prices {
		pricePtrs[i] = &prices[i]
	}

	return &types.QueryPricesResponse{Prices: pricePtrs}, nil
}
