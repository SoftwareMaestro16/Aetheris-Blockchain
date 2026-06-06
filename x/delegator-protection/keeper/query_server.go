package keeper

import (
	"context"

	"github.com/sovereign-l1/l1/x/delegator-protection/types"
	delegatorprotectionpb "github.com/sovereign-l1/l1/x/delegator-protection/types/delegatorprotectionpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ delegatorprotectionpb.QueryServer = Keeper{}

func (k Keeper) ProtectionFund(ctx context.Context, _ *delegatorprotectionpb.QueryProtectionFundRequest) (*delegatorprotectionpb.QueryProtectionFundResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryProtectionFund(state))
	return &delegatorprotectionpb.QueryProtectionFundResponse{FundJson: out}, err
}

func (k Keeper) ProtectionClaims(ctx context.Context, req *delegatorprotectionpb.QueryProtectionClaimsRequest) (*delegatorprotectionpb.QueryProtectionClaimsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryProtectionClaims(state, types.QueryProtectionClaimsRequest{Delegator: req.Delegator, Status: req.Status}))
	return &delegatorprotectionpb.QueryProtectionClaimsResponse{ClaimsJson: out}, err
}

func (k Keeper) DelegatorCompensation(ctx context.Context, req *delegatorprotectionpb.QueryDelegatorCompensationRequest) (*delegatorprotectionpb.QueryDelegatorCompensationResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryDelegatorCompensation(state, types.QueryDelegatorCompensationRequest{Delegator: req.Delegator}))
	return &delegatorprotectionpb.QueryDelegatorCompensationResponse{PayoutsJson: out}, err
}

func (k Keeper) ProtectionParams(ctx context.Context, _ *delegatorprotectionpb.QueryProtectionParamsRequest) (*delegatorprotectionpb.QueryProtectionParamsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryProtectionParams(state))
	return &delegatorprotectionpb.QueryProtectionParamsResponse{ParamsJson: out}, err
}
