package keeper

import (
	"context"

	"github.com/sovereign-l1/l1/x/performance/types"
	performancepb "github.com/sovereign-l1/l1/x/performance/types/performancepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ performancepb.QueryServer = Keeper{}

func (k Keeper) ValidatorPerformance(ctx context.Context, req *performancepb.QueryValidatorPerformanceRequest) (*performancepb.QueryValidatorPerformanceResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := types.QueryValidatorPerformanceOracle(state, types.QueryValidatorPerformanceRequest{Epoch: req.Epoch, ValidatorAddress: req.ValidatorAddress})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	out, err := mustJSON(res.Aggregate)
	return &performancepb.QueryValidatorPerformanceResponse{AggregateJson: out}, err
}

func (k Keeper) PerformanceEpoch(ctx context.Context, req *performancepb.QueryPerformanceEpochRequest) (*performancepb.QueryPerformanceEpochResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := types.QueryPerformanceEpochOracle(state, types.QueryPerformanceEpochRequest{Epoch: req.Epoch})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	out, err := mustJSON(res.Epoch)
	return &performancepb.QueryPerformanceEpochResponse{EpochJson: out}, err
}

func (k Keeper) PerformanceReports(ctx context.Context, req *performancepb.QueryPerformanceReportsRequest) (*performancepb.QueryPerformanceReportsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := types.QueryPerformanceReportsOracle(state, types.QueryPerformanceReportsRequest{Epoch: req.Epoch, ValidatorAddress: req.ValidatorAddress})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	out, err := mustJSON(res.Reports)
	return &performancepb.QueryPerformanceReportsResponse{ReportsJson: out}, err
}

func (k Keeper) PerformanceParams(ctx context.Context, _ *performancepb.QueryPerformanceParamsRequest) (*performancepb.QueryPerformanceParamsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryPerformanceParamsOracle(state).Params)
	return &performancepb.QueryPerformanceParamsResponse{ParamsJson: out}, err
}
