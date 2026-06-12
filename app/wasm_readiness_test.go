package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wasmconfig"
)

func TestDefaultAppDoesNotWireCosmWasmUntilFeatureGate(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

	require.NotContains(t, app.keys, wasmconfig.StoreKey)
	require.NotContains(t, genesis, wasmconfig.ModuleName)
}
