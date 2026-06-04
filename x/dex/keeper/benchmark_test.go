package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func BenchmarkDexCreatePoolsAndSwap(b *testing.B) {
	app := l1app.Setup(b, false)
	ctx := app.NewContext(false)
	trader := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(int64(maxBenchmarkIterations(b.N)*1_000+10_000)))[0]
	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		denom := benchmarkAssetDenom(i)
		b.StopTimer()
		fundAccount(b, app, ctx, trader, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_010)))
		b.StartTimer()

		createRes, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
			Creator: trader.String(),
			TokenA:  sdk.NewInt64Coin(denom, 1_000),
			TokenB:  sdk.NewInt64Coin(appparams.BaseDenom, 1_000),
		})
		require.NoError(b, err)

		_, err = msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
			Trader:        trader.String(),
			PoolId:        createRes.PoolId,
			TokenIn:       sdk.NewInt64Coin(denom, 10),
			TokenOutDenom: appparams.BaseDenom,
			MinAmountOut:  "1",
		})
		require.NoError(b, err)
	}
}

func benchmarkAssetDenom(i int) string {
	return fmt.Sprintf("benchasset%d", i)
}

func maxBenchmarkIterations(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
