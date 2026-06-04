package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/tests/bench"
)

func BenchmarkTokenFactoryInitGenesis(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("denoms_%d", size), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			admin := bench.GenesisWithLoad(b, 0, 0).Sender.String()
			gs := bench.TokenFactoryGenesis(admin, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				app, ctx = bench.InitializedApp(b)
				b.StartTimer()

				app.TokenFactoryKeeper.InitGenesis(ctx, *gs)
			}
		})
	}
}

func BenchmarkTokenFactoryExportGenesis(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("denoms_%d", size), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			admin := bench.GenesisWithLoad(b, 0, 0).Sender.String()
			bench.SeedTokenFactoryDenoms(b, app, ctx, admin, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				gs := app.TokenFactoryKeeper.ExportGenesis(ctx)
				require.Len(b, gs.Denoms, size)
			}
		})
	}
}

func BenchmarkTokenFactoryGetAllDenoms(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("denoms_%d", size), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			admin := bench.GenesisWithLoad(b, 0, 0).Sender.String()
			bench.SeedTokenFactoryDenoms(b, app, ctx, admin, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				denoms, err := app.TokenFactoryKeeper.GetAllDenoms(ctx)
				require.NoError(b, err)
				require.Len(b, denoms, size)
			}
		})
	}
}
