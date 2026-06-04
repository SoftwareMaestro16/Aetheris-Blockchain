package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/dex/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Pool(ctx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.PoolId == 0 {
		return nil, status.Error(codes.InvalidArgument, "pool_id must be positive")
	}
	pool, found, err := k.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, err
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
	pools, hasMore, err := k.GetPools(ctx, types.MaxQueryPools)
	if err != nil {
		return nil, err
	}
	if hasMore {
		return nil, status.Errorf(codes.ResourceExhausted, "too many pools; prototype query limit is %d", types.MaxQueryPools)
	}
	return &types.QueryPoolsResponse{Pools: pools}, nil
}
