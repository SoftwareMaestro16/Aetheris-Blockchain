#!/usr/bin/env bash
# TPS benchmark runner
# Runs the Go TPS benchmark and reports results.
# Usage: ./scripts/tps_test.sh [benchmark-time=5s] [txs-per-block=N]

set -euo pipefail

BENCH_TIME="${1:-5s}"
TXS_PER_BLOCK="${2:-100}"

echo "=== Aetra TPS Benchmark ==="
echo "Benchmark time:  $BENCH_TIME"
echo "Txs per block:   $TXS_PER_BLOCK"
echo ""

cd "$(dirname "$0")/.."

go test -bench=BenchmarkTPS -benchtime="$BENCH_TIME" -count=1 -run=^$ ./app/ 2>&1

echo ""
echo "=== Done ==="
