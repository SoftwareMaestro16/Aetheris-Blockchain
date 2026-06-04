package keeper_test

import (
	"fmt"
	"testing"

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

func TestPoolsQueryEnforcesPrototypeLimit(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	for i := 0; i < types.MaxQueryPools+1; i++ {
		poolID := uint64(i + 1)
		require.NoError(t, app.DexKeeper.SetPool(ctx, types.Pool{
			Id:          poolID,
			Denom0:      appparams.BaseDenom,
			Denom1:      fmt.Sprintf("queryasset%d", i),
			Reserve0:    "100",
			Reserve1:    "100",
			TotalShares: "100",
			LpDenom:     fmt.Sprintf("lp/%d", poolID),
		}))
	}

	_, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{})
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
}
