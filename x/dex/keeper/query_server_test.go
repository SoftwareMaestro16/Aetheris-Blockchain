package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func TestPoolsQueryReturnsEmptyState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	res, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{})
	require.NoError(t, err)
	require.Empty(t, res.Pools)
	require.NotNil(t, res.Pagination)
	require.Empty(t, res.Pagination.NextKey)
}

func TestPoolQueryReturnsPool(t *testing.T) {
	app, ctx, _, _, poolID := setupDexPool(t)

	res, err := app.DexKeeper.Pool(ctx, &types.QueryPoolRequest{PoolId: poolID})
	require.NoError(t, err)
	require.Equal(t, poolID, res.Pool.Id)
	require.Equal(t, "lp/1", res.Pool.LpDenom)
}

func TestPoolQueryErrorsAreGrpcStatusCompatible(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.DexKeeper.Pool(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.DexKeeper.Pool(ctx, &types.QueryPoolRequest{PoolId: 0})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.DexKeeper.Pool(ctx, &types.QueryPoolRequest{PoolId: 999})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestPoolsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.DexKeeper.Pools(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestPoolsQueryPaginatesWithNextKey(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	seedPools(t, app, ctx, 5)

	first, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &sdkquery.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, first.Pools, 2)
	require.Equal(t, uint64(1), first.Pools[0].Id)
	require.Equal(t, uint64(2), first.Pools[1].Id)
	require.NotEmpty(t, first.Pagination.NextKey)

	next, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &sdkquery.PageRequest{Limit: 2, Key: first.Pagination.NextKey},
	})
	require.NoError(t, err)
	require.Len(t, next.Pools, 2)
	require.Equal(t, uint64(3), next.Pools[0].Id)
	require.Equal(t, uint64(4), next.Pools[1].Id)
	require.NotEmpty(t, next.Pagination.NextKey)
}

func TestPoolsQueryDefaultLimitIsBoundedOnLargeState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	seedPools(t, app, ctx, types.MaxQueryPools+1)

	res, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Pools, types.DefaultQueryPools)
	require.NotEmpty(t, res.Pagination.NextKey)
}

func TestPoolsQueryRejectsInvalidPagination(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	cases := []sdkquery.PageRequest{
		{Limit: types.MaxQueryPools + 1},
		{Key: []byte("bad-key")},
		{Offset: 1},
		{CountTotal: true},
		{Reverse: true},
	}

	for _, tc := range cases {
		_, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{Pagination: &tc})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
}

func seedPools(t *testing.T, app *l1app.L1App, ctx sdk.Context, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		poolID := uint64(i + 1)
		require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
			Id:          poolID,
			Denom0:      appparams.BaseDenom,
			Denom1:      fmt.Sprintf("queryasset%03d", i),
			Reserve0:    "100",
			Reserve1:    "100",
			TotalShares: "100",
			LpDenom:     fmt.Sprintf("lp/%d", poolID),
		}))
	}
}
