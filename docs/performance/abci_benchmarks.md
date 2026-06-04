# ABCI And Module Performance Benchmarks

This branch adds non-production benchmark coverage for Orbitalis consensus-critical paths.

## Audit Summary

| Path | Bound | Chain-halt / DoS risk | Benchmark coverage |
| --- | --- | --- | --- |
| `PreBlocker`, `BeginBlocker`, `EndBlocker` | Expected O(module hook work). Current custom modules do not add block hooks. Core SDK modules may scan their own bounded queues. | Slow hooks can delay block finalization. | `BenchmarkABCIHooksWithLargeStores` |
| App genesis import | O(genesis object count + JSON size). | Large genesis can slow bootstrapping and validator recovery. | `BenchmarkLargeGenesisInit` |
| Full block processing | O(tx count + ante/msg execution). | Oversized blocks can delay finality; CometBFT consensus params and gas limits bound this in production. | `BenchmarkFullBlockProcessing` |
| `x/tokenfactory` `InitGenesis` | O(denom count). | Large factory-denom genesis increases startup time. | `BenchmarkTokenFactoryInitGenesis` |
| `x/tokenfactory` `ExportGenesis` / `GetAllDenoms` | O(denom count), store-size dependent. | Export/query path can be expensive; public queries should use pagination before production exposure at scale. | `BenchmarkTokenFactoryExportGenesis`, `BenchmarkTokenFactoryGetAllDenoms` |
| `x/dex` `InitGenesis` | O(pool count). | Large pool genesis increases startup time. | `BenchmarkDexInitGenesis` |
| `x/dex` `ExportGenesis` / `GetAllPools` | O(pool count), store-size dependent. | Export/query path can be expensive; public queries should use pagination before production exposure at scale. | `BenchmarkDexExportGenesis`, `BenchmarkDexGetAllPools` |
| `x/fees` ante denom policy | O(number of fee denoms per tx). v1 policy allows only `norb`, so normal tx cost is effectively O(1). | Malformed txs with large fee vectors are bounded by tx decoding/size limits and ante execution. | `BenchmarkFeeAnteHandlerChecks` |

## Local Commands

Fast compile and correctness gates:

```powershell
$env:PATH = "$PWD\.work\tools\bin;$PWD\.work\tools\go1.25.11\go\bin;$env:PATH"
$env:GOWORK = "off"
$env:GOFLAGS = "-p=1"
go test ./...
go vet ./...
buf lint
buf generate
```

Benchmark smoke run:

```powershell
go test -run "^$" -bench "Benchmark(FeeAnteHandlerChecks|TokenFactoryExportGenesis|DexExportGenesis)" -benchtime=1x ./x/fees/keeper ./x/tokenfactory/keeper ./x/dex/keeper
```

Full nightly-style run:

```powershell
go test -run "^$" -bench "Benchmark" -benchtime=1x ./app ./x/tokenfactory/keeper ./x/dex/keeper ./x/fees/keeper
```

The benchmark suite uses 1k, 10k, and 100k object scenarios where reasonable. Results should be kept in PR notes or CI artifacts, not committed as generated files.

## Local Baseline

Environment: Windows amd64, AMD Ryzen 5 5600G, Go `1.25.11`, `-benchtime=1x`, `GOFLAGS=-p=1`.

| Benchmark | Size | Time | Allocation |
| --- | ---: | ---: | ---: |
| `BenchmarkLargeGenesisInit` | 1k denoms + 1k pools | 65.6913 ms | 33.94 MB |
| `BenchmarkLargeGenesisInit` | 10k denoms + 10k pools | 205.3985 ms | 97.84 MB |
| `BenchmarkLargeGenesisInit` | 100k denoms + 100k pools | 1.9553483 s | 756.60 MB |
| `BenchmarkABCIHooksWithLargeStores` | 1k denoms + 1k pools | 5.5469 ms | 508.62 KB |
| `BenchmarkABCIHooksWithLargeStores` | 10k denoms + 10k pools | 3.2098 ms | 242.84 KB |
| `BenchmarkABCIHooksWithLargeStores` | 100k denoms + 100k pools | 859.6 us | 238.60 KB |
| `BenchmarkFullBlockProcessing` | 100 tx | 50.2786 ms | 8.20 MB |
| `BenchmarkFullBlockProcessing` | 1k tx | 469.9155 ms | 73.30 MB |
| `BenchmarkFullBlockProcessing` | 10k tx | 4.4258907 s | 727.03 MB |
| `BenchmarkFeeAnteHandlerChecks` | 1k checks | 841.9 us | 1.27 MB |
| `BenchmarkFeeAnteHandlerChecks` | 10k checks | 6.5802 ms | 12.68 MB |
| `BenchmarkFeeAnteHandlerChecks` | 100k checks | 77.6372 ms | 126.80 MB |
| `BenchmarkTokenFactoryInitGenesis` | 1k denoms | 965.9 us | 587.54 KB |
| `BenchmarkTokenFactoryInitGenesis` | 10k denoms | 5.4703 ms | 5.43 MB |
| `BenchmarkTokenFactoryInitGenesis` | 100k denoms | 79.1087 ms | 50.78 MB |
| `BenchmarkTokenFactoryExportGenesis` | 1k denoms | 794.5 us | 448.78 KB |
| `BenchmarkTokenFactoryExportGenesis` | 10k denoms | 7.5178 ms | 5.24 MB |
| `BenchmarkTokenFactoryExportGenesis` | 100k denoms | 113.7428 ms | 55.25 MB |
| `BenchmarkTokenFactoryGetAllDenoms` | 1k denoms | 988.7 us | 448.69 KB |
| `BenchmarkTokenFactoryGetAllDenoms` | 10k denoms | 7.2853 ms | 5.23 MB |
| `BenchmarkTokenFactoryGetAllDenoms` | 100k denoms | 97.3068 ms | 55.25 MB |
| `BenchmarkDexInitGenesis` | 1k pools | 584.9 us | 586.86 KB |
| `BenchmarkDexInitGenesis` | 10k pools | 5.19 ms | 5.43 MB |
| `BenchmarkDexInitGenesis` | 100k pools | 67.3591 ms | 50.78 MB |
| `BenchmarkDexExportGenesis` | 1k pools | 789.5 us | 615.41 KB |
| `BenchmarkDexExportGenesis` | 10k pools | 7.991 ms | 8.36 MB |
| `BenchmarkDexExportGenesis` | 100k pools | 116.3033 ms | 92.23 MB |
| `BenchmarkDexGetAllPools` | 1k pools | 750.1 us | 615.20 KB |
| `BenchmarkDexGetAllPools` | 10k pools | 7.716 ms | 8.36 MB |
| `BenchmarkDexGetAllPools` | 100k pools | 102.9842 ms | 92.23 MB |
