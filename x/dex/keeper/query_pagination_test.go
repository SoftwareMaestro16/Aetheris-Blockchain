package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func createPoolForQuery(t *testing.T, app *l1app.L1App, ctx sdk.Context, msgServer types.MsgServer, creator sdk.AccAddress, i int) uint64 {
	t.Helper()

	denomA := fmt.Sprintf("asset%03d", i)
	denomB := fmt.Sprintf("quote%03d", i)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin(denomA, 100), sdk.NewInt64Coin(denomB, 100)))
	res, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin(denomA, 10),
		TokenB:  sdk.NewInt64Coin(denomB, 10),
	})
	require.NoError(t, err)
	return res.PoolId
}

func setupDexQueryFixture(t *testing.T, count int) (*l1app.L1App, sdk.Context, types.MsgServer, sdk.AccAddress) {
	t.Helper()

	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	for i := range count {
		createPoolForQuery(t, app, ctx, msgServer, creator, i)
	}
	return app, ctx, msgServer, creator
}

func TestPoolsQueryRejectsNilRequest(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, 0)

	_, err := app.DexKeeper.Pools(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty request")
}

func TestPoolsQueryEmptyState(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, 0)

	res, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{})
	require.NoError(t, err)
	require.Empty(t, res.Pools)
	require.NotNil(t, res.Pagination)
	require.Empty(t, res.Pagination.NextKey)
}

func TestPoolsQueryPagination(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, 105)

	first, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{})
	require.NoError(t, err)
	require.Len(t, first.Pools, types.DefaultPoolsQueryLimit)
	require.NotNil(t, first.Pagination)
	require.NotEmpty(t, first.Pagination.NextKey)
	require.Equal(t, uint64(1), first.Pools[0].Id)
	require.Equal(t, uint64(100), first.Pools[99].Id)

	next, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, next.Pools, 5)
	require.Equal(t, uint64(101), next.Pools[0].Id)
	require.Empty(t, next.Pagination.NextKey)

	offset, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &querytypes.PageRequest{Offset: 100, Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, offset.Pools, 5)

	limited, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &querytypes.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Len(t, limited.Pools, 2)
}

func TestPoolsQueryInvalidPagination(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, 3)
	first, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &querytypes.PageRequest{Limit: 1},
	})
	require.NoError(t, err)
	require.NotEmpty(t, first.Pagination.NextKey)

	for name, req := range map[string]*types.QueryPoolsRequest{
		"key and offset": {Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Offset: 1, Limit: 1}},
		"bad key prefix": {Pagination: &querytypes.PageRequest{Key: []byte{0x09}, Limit: 1}},
		"count total":    {Pagination: &querytypes.PageRequest{CountTotal: true}},
		"large offset":   {Pagination: &querytypes.PageRequest{Offset: types.MaxPoolsQueryOffset + 1}},
		"reverse":        {Pagination: &querytypes.PageRequest{Reverse: true, Limit: 1}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := app.DexKeeper.Pools(ctx, req)
			require.Error(t, err)
		})
	}
}

func TestPoolsQueryCapsLimitForManyPools(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, int(types.MaxPoolsQueryLimit)+5)

	res, err := app.DexKeeper.Pools(ctx, &types.QueryPoolsRequest{
		Pagination: &querytypes.PageRequest{Limit: types.MaxPoolsQueryLimit + 100},
	})
	require.NoError(t, err)
	require.Len(t, res.Pools, int(types.MaxPoolsQueryLimit))
	require.NotEmpty(t, res.Pagination.NextKey)
}

func TestPoolByPairQuery(t *testing.T) {
	app, ctx, _, _ := setupDexQueryFixture(t, 1)

	_, err := app.DexKeeper.PoolByPair(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty request")

	res, err := app.DexKeeper.PoolByPair(ctx, &types.QueryPoolByPairRequest{
		DenomA: "quote000",
		DenomB: "asset000",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.Pool.Id)
	require.Equal(t, "asset000", res.Pool.Denom0)
	require.Equal(t, "quote000", res.Pool.Denom1)

	_, err = app.DexKeeper.PoolByPair(ctx, &types.QueryPoolByPairRequest{DenomA: "missinga", DenomB: "missingb"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool not found")

	_, err = app.DexKeeper.PoolByPair(ctx, &types.QueryPoolByPairRequest{DenomA: "asset000", DenomB: "asset000"})
	require.Error(t, err)

	_, err = app.DexKeeper.PoolByPair(ctx, &types.QueryPoolByPairRequest{DenomA: "bad denom", DenomB: "asset000"})
	require.Error(t, err)
}

func TestCreatePoolRejectsDuplicatePair(t *testing.T) {
	app, ctx, msgServer, creator := setupDexQueryFixture(t, 1)
	fundAccount(t, app, ctx, creator, sdk.NewCoins(sdk.NewInt64Coin("asset000", 100), sdk.NewInt64Coin("quote000", 100)))

	_, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("quote000", 10),
		TokenB:  sdk.NewInt64Coin("asset000", 10),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pool already exists")
}
