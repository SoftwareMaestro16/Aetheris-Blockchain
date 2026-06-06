package keeper

import (
	"context"

	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ reputationpb.QueryServer = Keeper{}

func (k Keeper) ValidatorReputation(ctx context.Context, req *reputationpb.QueryValidatorReputationRequest) (*reputationpb.QueryValidatorReputationResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	addr, err := parseAddress(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	record, found := types.QueryValidatorReputation(state, addr)
	if !found {
		return nil, status.Error(codes.NotFound, "reputation record not found")
	}
	out, err := mustJSON(record)
	return &reputationpb.QueryValidatorReputationResponse{RecordJson: out}, err
}

func (k Keeper) ReporterReputation(ctx context.Context, req *reputationpb.QueryReporterReputationRequest) (*reputationpb.QueryReporterReputationResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	addr, err := parseAddress(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	record, found := types.QueryReporterReputation(state, addr)
	if !found {
		return nil, status.Error(codes.NotFound, "reputation record not found")
	}
	out, err := mustJSON(record)
	return &reputationpb.QueryReporterReputationResponse{RecordJson: out}, err
}

func (k Keeper) ReputationHistory(ctx context.Context, req *reputationpb.QueryReputationHistoryRequest) (*reputationpb.QueryReputationHistoryResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	addr, err := parseAddress(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	snapshots, events, err := types.QueryReputationHistory(state, types.ReputationHistoryQuery{SubjectType: req.SubjectType, Subject: addr, Limit: req.Limit})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sj, err := mustJSON(snapshots)
	if err != nil {
		return nil, err
	}
	ej, err := mustJSON(events)
	if err != nil {
		return nil, err
	}
	return &reputationpb.QueryReputationHistoryResponse{SnapshotsJson: sj, EventsJson: ej}, nil
}

func (k Keeper) ReputationParams(ctx context.Context, _ *reputationpb.QueryReputationParamsRequest) (*reputationpb.QueryReputationParamsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryReputationParams(state))
	return &reputationpb.QueryReputationParamsResponse{ParamsJson: out}, err
}
