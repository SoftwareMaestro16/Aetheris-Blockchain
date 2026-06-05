package wasmconfig

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	governanceAddr = "0:0000000000000000000000000000000000000000000000000000000000000001"
	uploaderAddr   = "0:0000000000000000000000000000000000000000000000000000000000000002"
	ownerAddr      = "0:0000000000000000000000000000000000000000000000000000000000000003"
	contractAddr   = "0:0000000000000000000000000000000000000000000000000000000000000004"
	attackerAddr   = "0:0000000000000000000000000000000000000000000000000000000000000005"
)

func TestDefaultPolicyIsDisabledAndPinnedToCompatibleWasmd(t *testing.T) {
	policy := DefaultPolicy()

	require.False(t, policy.Enabled)
	require.Equal(t, "v0.70.2", RecommendedWasmdVersion)
	require.Equal(t, "v3.0.6", RecommendedWasmVMVersion)
	require.Equal(t, "v0.54", RecommendedSDKMinor)
	require.Equal(t, UploadPermissionGovernanceOnly, policy.UploadPermission)
	require.Equal(t, InstantiatePermissionCodeOwnerOnly, policy.InstantiatePermission)
	require.NoError(t, policy.Validate())
	require.ErrorContains(t, CanUpload(governanceAddr, policy), "disabled by feature gate")
}

func TestPolicyRejectsUnsafeLimits(t *testing.T) {
	policy := enabledPolicy()

	tooLargeCode := policy
	tooLargeCode.MaxContractSizeBytes = DefaultMaxContractSizeBytes + 1
	require.ErrorContains(t, tooLargeCode.Validate(), "max contract size")

	tooLargeQueryGas := policy
	tooLargeQueryGas.SmartQueryGasLimit = maxSmartQueryGasLimit + 1
	require.ErrorContains(t, tooLargeQueryGas.Validate(), "smart query gas")

	unbenchmarkedGasMultiplier := policy
	unbenchmarkedGasMultiplier.GasMultiplier = DefaultGasMultiplier + 1
	require.ErrorContains(t, unbenchmarkedGasMultiplier.Validate(), "gas multiplier")

	tooMuchCache := policy
	tooMuchCache.MemoryCacheSizeMiB = maxMemoryCacheSizeMiB + 1
	require.ErrorContains(t, tooMuchCache.Validate(), "memory cache")
}

func TestAllowlistRejectsMalformedEmptyAndZeroAddress(t *testing.T) {
	policy := enabledPolicy()
	policy.UploadPermission = UploadPermissionAllowlist

	policy.UploadAllowlist = nil
	require.ErrorContains(t, policy.Validate(), "allowlist must not be empty")

	policy.UploadAllowlist = []string{"orb1malformed"}
	require.Error(t, policy.Validate())

	policy.UploadAllowlist = []string{addressing.ZeroRawAddress}
	require.ErrorContains(t, policy.Validate(), "must not be zero address")
}

func TestGovernanceOnlyUploadRequiresAuthority(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, CanUpload(governanceAddr, policy))
	require.ErrorContains(t, CanUpload(uploaderAddr, policy), "requires governance authority")

	policy.GovernanceAuthority = addressing.ZeroRawAddress
	require.ErrorContains(t, CanUpload(governanceAddr, policy), "must not be zero address")
}

func TestAllowlistUploadInstantiateExecuteMigrateLifecycle(t *testing.T) {
	policy := enabledPolicy()
	policy.UploadPermission = UploadPermissionAllowlist
	policy.InstantiatePermission = InstantiatePermissionEverybody
	policy.UploadAllowlist = []string{uploaderAddr}
	require.NoError(t, policy.Validate())

	require.NoError(t, CanUpload(uploaderAddr, policy))
	require.NoError(t, CanInstantiate(ownerAddr, uploaderAddr, policy))
	require.NoError(t, CanExecute(ownerAddr, contractAddr, policy))
	require.NoError(t, CanMigrate(ownerAddr, ownerAddr, policy))
}

func TestInstantiateOwnerOnlyAndMigrationRejectAdminTakeover(t *testing.T) {
	policy := enabledPolicy()

	require.NoError(t, CanInstantiate(ownerAddr, ownerAddr, policy))
	require.ErrorContains(t, CanInstantiate(attackerAddr, ownerAddr, policy), "requires code owner")
	require.ErrorContains(t, CanMigrate(attackerAddr, ownerAddr, policy), "requires contract admin")
	require.ErrorContains(t, CanMigrate(ownerAddr, "", policy), "empty address string")
	require.ErrorContains(t, CanMigrate(ownerAddr, addressing.ZeroUserFriendly, policy), "must not be zero address")
}

func enabledPolicy() Policy {
	policy := DefaultPolicy()
	policy.Enabled = true
	policy.GovernanceAuthority = governanceAddr
	return policy
}
