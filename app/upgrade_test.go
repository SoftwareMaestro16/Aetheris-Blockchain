package app

import (
	"encoding/json"
	"testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/require"

	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestModuleVersionMapIncludesPrototypeModules(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	stored, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	current := app.ModuleManager.GetVersionMap()

	for moduleName, version := range current {
		require.NotZero(t, version, moduleName)
		require.Equal(t, version, stored[moduleName], moduleName)
	}
	require.Equal(t, uint64(1), stored[tokenfactorytypes.ModuleName])
	require.Equal(t, uint64(1), stored[dextypes.ModuleName])
	require.Equal(t, uint64(1), stored[feestypes.ModuleName])
}

func TestNoOpUpgradeDryRunAndExport(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	before, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.True(t, app.UpgradeKeeper.HasHandler(UpgradeName))

	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgradetypes.Plan{
		Name:   UpgradeName,
		Height: ctx.BlockHeight(),
	}))

	after, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.Equal(t, app.ModuleManager.GetVersionMap(), after)
	require.Equal(t, before[tokenfactorytypes.ModuleName], after[tokenfactorytypes.ModuleName])
	require.Equal(t, before[dextypes.ModuleName], after[dextypes.ModuleName])
	require.Equal(t, before[feestypes.ModuleName], after[feestypes.ModuleName])

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))
}

func TestUpgradeVersionMapValidationRejectsMissingOrFutureModuleVersion(t *testing.T) {
	app := Setup(t, false)
	current := app.ModuleManager.GetVersionMap()

	missing := cloneVersionMap(current)
	delete(missing, tokenfactorytypes.ModuleName)
	require.ErrorContains(t, ValidateUpgradeVersionMap(missing, current), "missing module version")

	future := cloneVersionMap(current)
	future[dextypes.ModuleName] = current[dextypes.ModuleName] + 1
	require.ErrorContains(t, ValidateUpgradeVersionMap(future, current), "newer than current version")

	allowedNew := cloneVersionMap(current)
	delete(allowedNew, feestypes.ModuleName)
	require.NoError(t, ValidateUpgradeVersionMap(allowedNew, current, feestypes.ModuleName))
}

func cloneVersionMap(versionMap map[string]uint64) map[string]uint64 {
	out := make(map[string]uint64, len(versionMap))
	for moduleName, version := range versionMap {
		out[moduleName] = version
	}
	return out
}
