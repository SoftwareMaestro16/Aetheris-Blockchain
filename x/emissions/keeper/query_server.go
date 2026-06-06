package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/emissions/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) EmissionsParams(ctx context.Context, req *types.QueryEmissionsParamsRequest) (*types.QueryEmissionsParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryEmissionsParamsResponse{Params: params}, nil
}

func (k Keeper) CurrentInflation(ctx context.Context, req *types.QueryCurrentInflationRequest) (*types.QueryCurrentInflationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryCurrentInflationResponse{CurrentInflationBps: params.CurrentInflationBps}, nil
}

func (k Keeper) EmissionEpoch(ctx context.Context, req *types.QueryEmissionEpochRequest) (*types.QueryEmissionEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	record, found, err := k.GetEmissionEpoch(ctx, req.Epoch)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrNotFound.Error())
	}
	return &types.QueryEmissionEpochResponse{EmissionEpoch: record}, nil
}

func (k Keeper) DistributionWeights(ctx context.Context, req *types.QueryDistributionWeightsRequest) (*types.QueryDistributionWeightsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryDistributionWeightsResponse{Weights: params.DistributionWeights}, nil
}
