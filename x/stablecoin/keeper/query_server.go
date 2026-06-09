package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	k Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{k: k}
}

func (q queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Params must be a pointer in the response
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (q queryServer) TotalMinted(ctx context.Context, req *types.QueryTotalMintedRequest) (*types.QueryTotalMintedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	totalMinted, err := q.k.GetTotalMinted(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalMintedResponse{TotalMinted: totalMinted.String()}, nil
}

func (q queryServer) TotalBurned(ctx context.Context, req *types.QueryTotalBurnedRequest) (*types.QueryTotalBurnedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	totalBurned, err := q.k.GetTotalBurned(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTotalBurnedResponse{TotalBurned: totalBurned.String()}, nil
}

func (q queryServer) CirculatingSupply(ctx context.Context, req *types.QueryCirculatingSupplyRequest) (*types.QueryCirculatingSupplyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	circulating, err := q.k.GetCirculatingSupply(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryCirculatingSupplyResponse{CirculatingSupply: circulating.String()}, nil
}
