package upgrades

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/module"
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
