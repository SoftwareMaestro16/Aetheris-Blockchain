package app_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/tests/bench"
)

func BenchmarkLargeGenesisInit(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("denoms_%d_pools_%d", size, size), func(b *testing.B) {
			fixture := bench.GenesisWithLoad(b, size, size)
			b.ReportAllocs()
			b.SetBytes(int64(len(fixture.GenesisBytes)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = bench.InitApp(b, fixture.GenesisBytes)
			}
		})
	}
}

func BenchmarkABCIHooksWithLargeStores(b *testing.B) {
	for _, size := range bench.ObjectSizes {
		b.Run(fmt.Sprintf("denoms_%d_pools_%d", size, size), func(b *testing.B) {
			fixture := bench.GenesisWithLoad(b, size, size)
			l1 := bench.InitApp(b, fixture.GenesisBytes)
			baseCtx := l1.NewContext(false)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ctx, _ := baseCtx.CacheContext()
				_, err := l1.PreBlocker(ctx, &abci.RequestFinalizeBlock{Height: int64(i + 1)})
				require.NoError(b, err)
				_, err = l1.BeginBlocker(ctx)
				require.NoError(b, err)
				_, err = l1.EndBlocker(ctx)
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkFullBlockProcessing(b *testing.B) {
	for _, txCount := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("txs_%d", txCount), func(b *testing.B) {
			fixture := bench.GenesisWithLoad(b, 0, 0)
			txs := bench.MsgSendTxs(b, fixture.App, fixture.SenderPriv, fixture.Sender, fixture.Recipient, txCount)
			var totalTxBytes int64
			for _, tx := range txs {
				totalTxBytes += int64(len(tx))
			}
			b.ReportAllocs()
			b.SetBytes(totalTxBytes)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				l1 := bench.InitApp(b, fixture.GenesisBytes)
				b.StartTimer()

				resp, err := l1.FinalizeBlock(&abci.RequestFinalizeBlock{
					Height:             1,
					Hash:               []byte("bench-full-block"),
					NextValidatorsHash: fixture.ValSetHash,
					Txs:                txs,
				})
				require.NoError(b, err)
				require.Len(b, resp.TxResults, txCount)
				for _, txResult := range resp.TxResults {
					require.Equal(b, uint32(0), txResult.Code, txResult.Log)
				}
				_, err = l1.Commit()
				require.NoError(b, err)
			}
		})
	}
}
