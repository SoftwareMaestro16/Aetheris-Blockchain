package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	"github.com/sovereign-l1/l1/x/dex/types"
)

func benchmarkDexPool(b *testing.B) (*l1app.L1App, sdk.Context, types.MsgServer, sdk.AccAddress, uint64) {
	b.Helper()

	app := l1app.Setup(b, false)
	ctx := app.NewContext(false)
	creator := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000_000_000_000))[0]
	requireBenchmarkNoError(b, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000_000_000_000_000))))
	requireBenchmarkNoError(b, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, creator, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1_000_000_000_000_000))))

	msgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)
	res, err := msgServer.CreatePool(ctx, &types.MsgCreatePool{
		Creator: creator.String(),
		TokenA:  sdk.NewInt64Coin("uatom", 500_000_000_000_000),
		TokenB:  sdk.NewInt64Coin(appparams.BaseDenom, 500_000_000_000_000),
	})
	requireBenchmarkNoError(b, err)
	return app, ctx, msgServer, creator, res.PoolId
}

func BenchmarkDexSwapExactAmountIn(b *testing.B) {
	_, ctx, msgServer, trader, poolID := benchmarkDexPool(b)
	msg := &types.MsgSwapExactAmountIn{
		Trader:        trader.String(),
		PoolId:        poolID,
		TokenIn:       sdk.NewInt64Coin("uatom", 1_000_000),
		TokenOutDenom: appparams.BaseDenom,
		MinAmountOut:  "1",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := msgServer.SwapExactAmountIn(ctx, msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDexSetAndPagePools(b *testing.B) {
	app, ctx, _, _, _ := benchmarkDexPool(b)
	for i := uint64(2); i <= 100; i++ {
		pool := types.Pool{
			Id:          i,
			Denom0:      fmt.Sprintf("factory/4:0000000000000000000000001111111111111111111111111111111111111111/t%03d", i),
			Denom1:      appparams.BaseDenom,
			Reserve0:    "1000",
			Reserve1:    "1000",
			TotalShares: "1000",
			LpDenom:     fmt.Sprintf("lp/%d", i),
		}
		requireBenchmarkNoError(b, app.DexKeeper.SetPool(ctx, pool))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := app.DexKeeper.GetPoolsPage(ctx, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func requireBenchmarkNoError(b *testing.B, err error) {
	b.Helper()
	if err != nil {
		b.Fatal(err)
	}
}
