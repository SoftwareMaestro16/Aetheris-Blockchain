package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func BenchmarkValidateFeeCoinsAllowedNaet(b *testing.B) {
	params := types.DefaultParams()
	fees := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := types.ValidateFeeCoins(params, fees, true); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateFeeCoinsRejectsNonNaet(b *testing.B) {
	params := types.DefaultParams()
	fees := sdk.NewCoins(sdk.NewInt64Coin("factory/4:0000000000000000000000001111111111111111111111111111111111111111/usdt", 1))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := types.ValidateFeeCoins(params, fees, true); err == nil {
			b.Fatal("expected non-naet fee rejection")
		}
	}
}
