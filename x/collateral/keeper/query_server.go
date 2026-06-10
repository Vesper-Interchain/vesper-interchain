package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	k Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{k: k}
}

func (q queryServer) Position(ctx context.Context, req *types.QueryPositionRequest) (*types.QueryPositionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	position, err := q.k.GetPosition(sdkCtx, req.Owner)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryPositionResponse{Position: &position}, nil
}

func (q queryServer) CollateralRatio(ctx context.Context, req *types.QueryCollateralRatioRequest) (*types.QueryCollateralRatioResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	position, err := q.k.GetPosition(sdkCtx, req.Owner)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	collateralRatio, err := q.k.GetCollateralRatio(sdkCtx, position)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	collateralValueUSD, err := q.k.GetCollateralValueUSD(sdkCtx, position.CollateralAmount, position.CollateralDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	debtValueUSD, err := math.LegacyNewDecFromStr(position.DebtAmount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	debtValueUSD = debtValueUSD.QuoInt64(1_000_000)

	return &types.QueryCollateralRatioResponse{
		CollateralRatio:    collateralRatio.String(),
		CollateralValueUsd: collateralValueUSD.String(),
		DebtValueUsd:       debtValueUSD.String(),
	}, nil
}

func (q queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := q.k.Params.Get(sdkCtx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: &params}, nil
}
