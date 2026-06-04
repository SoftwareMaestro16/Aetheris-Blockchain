package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/dex/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Pool(ctx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	pool, found, err := k.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	return &types.QueryPoolResponse{Pool: pool}, nil
}

func (k Keeper) Pools(ctx context.Context, req *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	pools, pagination, err := k.GetPoolsPaginated(ctx, req.Pagination)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrInvalidPagination) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPoolsResponse{Pools: pools, Pagination: pagination}, nil
}

func (k Keeper) PoolByPair(ctx context.Context, req *types.QueryPoolByPairRequest) (*types.QueryPoolByPairResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	denom0, denom1, err := canonicalDenomPair(req.DenomA, req.DenomB)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	poolID, found, err := k.GetPoolIDByPair(ctx, denom0, denom1)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	pool, found, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.Internal, "pool index points to missing pool")
	}
	return &types.QueryPoolByPairResponse{Pool: pool}, nil
}
