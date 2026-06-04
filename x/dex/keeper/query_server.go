package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/dex/types"
	queryutil "github.com/sovereign-l1/l1/x/internal/query"
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
	pools, page, err := k.GetPoolsPage(ctx, req.Pagination)
	if err != nil {
		if errors.Is(err, queryutil.ErrInvalidPagination) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pagination: %v", err)
		}
		return nil, err
	}
	return &types.QueryPoolsResponse{Pools: pools, Pagination: page}, nil
}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}
