package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/tests/bench"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func BenchmarkFeeAnteHandlerChecks(b *testing.B) {
	for _, checks := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("checks_%d", checks), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			handler := app.FeesKeeper.AnteHandlerDecorator(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				return ctx, nil
			})
			tx := feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))}
			b.ReportAllocs()
			b.ReportMetric(float64(checks), "checks/op")
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for j := 0; j < checks; j++ {
					_, err := handler(ctx, tx, false)
					require.NoError(b, err)
				}
			}
		})
	}
}
