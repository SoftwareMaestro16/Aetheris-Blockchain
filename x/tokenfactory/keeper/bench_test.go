package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	l1app "github.com/sovereign-l1/l1/app"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func benchmarkTokenfactory(b *testing.B) (*l1app.L1App, sdk.Context, types.MsgServer, sdk.AccAddress) {
	b.Helper()

	app := l1app.Setup(b, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	return app, ctx, msgServer, admin
}

func BenchmarkTokenfactoryCreateDenom(b *testing.B) {
	_, ctx, msgServer, admin := benchmarkTokenfactory(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
			Creator:  admin.String(),
			Subdenom: fmt.Sprintf("bench%08d", i),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenfactoryMint(b *testing.B) {
	_, ctx, msgServer, admin := benchmarkTokenfactory(b)
	res, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "mintbench",
	})
	if err != nil {
		b.Fatal(err)
	}
	msg := &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(res.NewTokenDenom, 1),
		MintToAddress: admin.String(),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := msgServer.Mint(ctx, msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenfactoryGetDenomsPage(b *testing.B) {
	app, ctx, msgServer, admin := benchmarkTokenfactory(b)
	for i := 0; i < 100; i++ {
		_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
			Creator:  admin.String(),
			Subdenom: fmt.Sprintf("page%03d", i),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := app.TokenFactoryKeeper.GetDenomsPage(ctx, nil); err != nil {
			b.Fatal(err)
		}
	}
}
