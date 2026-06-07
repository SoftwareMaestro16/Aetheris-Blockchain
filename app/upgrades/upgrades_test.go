package upgrades

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/module"

	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

func TestValidateVersionMapRejectsMissingUnexpectedModule(t *testing.T) {
	err := ValidateVersionMap(module.VersionMap{"auth": 1}, module.VersionMap{"auth": 1, "bank": 1})
	require.ErrorContains(t, err, "missing module version for bank")
}

func TestValidateVersionMapAllowsExplicitNewModule(t *testing.T) {
	err := ValidateVersionMap(module.VersionMap{"auth": 1}, module.VersionMap{"auth": 1, "bank": 1}, "bank")
	require.NoError(t, err)
}

func TestValidateVersionMapRejectsDowngradeSourceNewerThanCurrent(t *testing.T) {
	err := ValidateVersionMap(module.VersionMap{"auth": 2}, module.VersionMap{"auth": 1})
	require.ErrorContains(t, err, "newer than current")
}

func TestNativeAccountVersionUpgradeHandlerAllowsNativeAccountModule(t *testing.T) {
	plan := NativeAccountVersionUpgradePlan()
	fromVM := module.VersionMap{"auth": 1}
	currentVM := module.VersionMap{"auth": 1, plan.ModuleName: 2}

	require.NoError(t, ValidateNativeAccountVersionUpgradeHandler(fromVM, currentVM))

	require.NoError(t, nativeaccounttypes.ValidateNativeAccountVersionUpgradePlan(plan))
	require.False(t, plan.RequiresFullBlockScan)
	require.True(t, plan.LazyMigrationEnabled)
}

func TestNativeAccountVersionUpgradeHandlerStillRejectsUnexpectedModules(t *testing.T) {
	plan := NativeAccountVersionUpgradePlan()
	fromVM := module.VersionMap{"auth": 1}
	currentVM := module.VersionMap{"auth": 1, plan.ModuleName: 2, "unexpected": 1}

	err := ValidateNativeAccountVersionUpgradeHandler(fromVM, currentVM)
	require.ErrorContains(t, err, "missing module version for unexpected")
}
