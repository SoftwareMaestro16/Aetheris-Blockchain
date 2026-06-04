package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Denom(ctx context.Context, req *types.QueryDenomRequest) (*types.QueryDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	meta, found, err := k.GetDenom(ctx, req.Denom)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrInvalidAddress, types.ErrInvalidDenom) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
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
	denoms, pagination, err := k.GetDenomsPaginated(ctx, req.Pagination)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrInvalidPagination) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryDenomsResponse{Denoms: denoms, Pagination: pagination}, nil
}
