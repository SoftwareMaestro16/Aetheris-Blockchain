package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Denom(ctx context.Context, req *types.QueryDenomRequest) (*types.QueryDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid denom: %v", err)
	}
	meta, found, err := k.GetDenom(ctx, req.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "denom not found")
	}
	return &types.QueryDenomResponse{Metadata: meta}, nil
}

func (k Keeper) Denoms(ctx context.Context, req *types.QueryDenomsRequest) (*types.QueryDenomsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	denoms, hasMore, err := k.GetDenoms(ctx, types.MaxQueryDenoms)
	if err != nil {
		return nil, err
	}
	if hasMore {
		return nil, status.Errorf(codes.ResourceExhausted, "too many denoms; prototype query limit is %d", types.MaxQueryDenoms)
	}
	return &types.QueryDenomsResponse{Denoms: denoms}, nil
}
