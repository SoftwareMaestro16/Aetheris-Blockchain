package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	queryutil "github.com/sovereign-l1/l1/x/internal/query"
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
	denoms, page, err := k.GetDenomsPage(ctx, req.Pagination)
	if err != nil {
		if errors.Is(err, queryutil.ErrInvalidPagination) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pagination: %v", err)
		}
		return nil, err
	}
	return &types.QueryDenomsResponse{Denoms: denoms, Pagination: page}, nil
}
