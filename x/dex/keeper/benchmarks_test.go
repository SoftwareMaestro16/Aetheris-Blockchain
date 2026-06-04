package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/tests/bench"
)

func BenchmarkDexInitGenesis(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("pools_%d", size), func(b *testing.B) {
			gs := bench.DexGenesis(size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				app, ctx := bench.InitializedApp(b)
				b.StartTimer()

				app.DexKeeper.InitGenesis(ctx, *gs)
			}
		})
	}
}

func BenchmarkDexExportGenesis(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("pools_%d", size), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			bench.SeedDexPools(b, app, ctx, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				gs := app.DexKeeper.ExportGenesis(ctx)
				require.Len(b, gs.Pools, size)
			}
		})
	}
}

func BenchmarkDexGetAllPools(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("pools_%d", size), func(b *testing.B) {
			app, ctx := bench.InitializedApp(b)
			bench.SeedDexPools(b, app, ctx, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				pools, err := app.DexKeeper.GetAllPools(ctx)
				require.NoError(b, err)
				require.Len(b, pools, size)
			}
		})
	}
}
