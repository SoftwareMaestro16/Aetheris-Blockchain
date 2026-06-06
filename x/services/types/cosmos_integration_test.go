package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultCosmosSDKServiceIntegrationManifest(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.RequiredModules, 6)
	require.Len(t, manifest.OptionalModules, 6)
	require.Len(t, manifest.KeeperBoundaries, 6)
	require.True(t, manifest.Rules.StoreV2IsolatedPrefixes)
	require.Equal(t, "messages", manifest.Rules.CrossZoneOperationMode)
	require.Equal(t, "bank_or_financial_zone", manifest.Rules.PaymentSettlementIntegration)
	require.Equal(t, ServiceModuleIdentity, manifest.Rules.IdentityAuthorizationModule)
	require.NotEmpty(t, manifest.ManifestHash)

	required := map[string]struct{}{}
	for _, module := range manifest.RequiredModules {
		required[module.ModulePath] = struct{}{}
		require.Equal(t, ServiceIntegrationModuleRequired, module.Kind)
	}
	for _, modulePath := range requiredServiceIntegrationModules() {
		_, found := required[modulePath]
		require.True(t, found, modulePath)
	}
}

func TestCosmosSDKServiceIntegrationKeeperBoundariesUseIsolatedStoreV2Prefixes(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	owners := map[string]ServiceKeeperBoundaryName{}
	for _, boundary := range manifest.KeeperBoundaries {
		require.NoError(t, boundary.Validate())
		for _, prefix := range boundary.StorePrefixes {
			require.True(t, IsServiceStoreKey(prefix+"/probe"), prefix)
			owners[prefix] = boundary.KeeperName
		}
	}
	require.Equal(t, PaymentKeeperBoundary, owners[ServicePaymentModelPrefix])
	require.Equal(t, ReceiptKeeperBoundary, owners[ServiceStorePrefix+"receipts"])
	require.Equal(t, InterfaceKeeperBoundary, owners[ServiceStorePrefix+"interfaces"])
}

func TestCosmosSDKServiceIntegrationRejectsMissingRequiredModule(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.RequiredModules = manifest.RequiredModules[1:]
	manifest.ManifestHash = ComputeCosmosSDKServiceIntegrationManifestHash(manifest)
	require.ErrorContains(t, manifest.Validate(), "required")
}

func TestCosmosSDKServiceIntegrationRejectsCrossKeeperPrefixOverlap(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	for i := range manifest.KeeperBoundaries {
		if manifest.KeeperBoundaries[i].KeeperName == PaymentKeeperBoundary {
			manifest.KeeperBoundaries[i].StorePrefixes[0] = ServiceStorePrefix + "descriptors"
			manifest.KeeperBoundaries[i].StorePrefixes = sortedStrings(manifest.KeeperBoundaries[i].StorePrefixes)
			manifest.KeeperBoundaries[i].BoundaryHash = ComputeServiceKeeperBoundaryHash(manifest.KeeperBoundaries[i])
		}
	}
	manifest.ManifestHash = ComputeCosmosSDKServiceIntegrationManifestHash(manifest)
	require.ErrorContains(t, manifest.Validate(), "prefix")
}

func TestCosmosSDKServiceIntegrationRejectsUnsafeIntegrationRules(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.Rules.CrossZoneOperationMode = "direct_writes"
	rehashCosmosIntegrationManifest(&manifest)
	require.ErrorContains(t, manifest.Validate(), "cross-zone")

	manifest, err = DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.Rules.PaymentSettlementIntegration = "local_keeper_only"
	rehashCosmosIntegrationManifest(&manifest)
	require.ErrorContains(t, manifest.Validate(), "payment settlement")

	manifest, err = DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.Rules.IdentityAuthorizationModule = "x/accounts"
	rehashCosmosIntegrationManifest(&manifest)
	require.ErrorContains(t, manifest.Validate(), "x/identity")

	manifest, err = DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.Rules.ContractExecutionIntegration = "direct_contract_write"
	rehashCosmosIntegrationManifest(&manifest)
	require.ErrorContains(t, manifest.Validate(), "contract module interface")
}

func TestCosmosSDKServiceIntegrationRejectsManifestHashMismatch(t *testing.T) {
	manifest, err := DefaultCosmosSDKServiceIntegrationManifest()
	require.NoError(t, err)
	manifest.ManifestHash = servicesHashParts("wrong/manifest")
	require.ErrorContains(t, manifest.Validate(), "hash mismatch")
}

func rehashCosmosIntegrationManifest(manifest *CosmosSDKServiceIntegrationManifest) {
	manifest.Rules.RulesHash = ComputeServiceKeeperIntegrationRulesHash(manifest.Rules)
	manifest.ManifestHash = ComputeCosmosSDKServiceIntegrationManifestHash(*manifest)
}
