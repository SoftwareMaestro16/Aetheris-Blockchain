package keeper

import (
	"context"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ reputationpb.QueryServer = Keeper{}

// ValidatorReputation returns the validator score (legacy query — kept for backward compat).
func (k Keeper) ValidatorReputation(ctx context.Context, req *reputationpb.QueryValidatorReputationRequest) (*reputationpb.QueryValidatorReputationResponse, error) {
	addr, err := parseAddress(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userAddr := aetraaddress.FormatAccAddress(addr)
	vs, err := k.GetValidatorReputation(ctx, userAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(vs)
	return &reputationpb.QueryValidatorReputationResponse{RecordJson: out}, err
}

// ReporterReputation is deprecated — returns not-found for all requests.
func (k Keeper) ReporterReputation(_ context.Context, _ *reputationpb.QueryReporterReputationRequest) (*reputationpb.QueryReporterReputationResponse, error) {
	return nil, status.Error(codes.NotFound, "reporter reputation is removed; use identity reputation for accounts, validator score for validators")
}

// ReputationHistory is deprecated — returns empty results.
func (k Keeper) ReputationHistory(_ context.Context, _ *reputationpb.QueryReputationHistoryRequest) (*reputationpb.QueryReputationHistoryResponse, error) {
	return &reputationpb.QueryReputationHistoryResponse{SnapshotsJson: "[]", EventsJson: "[]"}, nil
}

// ReputationParams returns the current params.
func (k Keeper) ReputationParams(ctx context.Context, _ *reputationpb.QueryReputationParamsRequest) (*reputationpb.QueryReputationParamsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(state.Params)
	return &reputationpb.QueryReputationParamsResponse{ParamsJson: out}, err
}

// IdentityReputationQuery returns the unified identity reputation.
func (k Keeper) IdentityReputationQuery(ctx context.Context, req *types.QueryIdentityReputationRequest) (*types.QueryIdentityReputationResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	account := req.Account
	id, found := types.FindIdentity(state, account)
	if !found {
		id = *types.NewIdentityReputation(account)
	}
	vs, _ := types.FindValidatorScore(state, account)
	sts, _ := types.FindServiceTrustScore(state, account)
	response := types.QueryIdentityReputation(&id, &vs, &sts)
	return &response, nil
}

// ValidatorScoreQuery returns the validator score.
func (k Keeper) ValidatorScoreQuery(ctx context.Context, addr string) (*types.ValidatorScore, error) {
	return k.GetValidatorReputation(ctx, addr)
}

// ServiceTrustScoreQuery returns the service trust score.
func (k Keeper) ServiceTrustScoreQuery(ctx context.Context, addr string) (*types.ServiceTrustScore, error) {
	return k.GetServiceTrustScore(ctx, addr)
}
