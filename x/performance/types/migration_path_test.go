package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationPhase0ReadinessPassesBaselineHardening(t *testing.T) {
	report := BuildMigrationPhase0Readiness(validMigrationPhase0Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
	require.NotEmpty(t, report.ReportHash)
}

func TestMigrationPhase0ReadinessFailsUnreproducibleStateMissingInvariantAndUnsafePrefix(t *testing.T) {
	input := validMigrationPhase0Input()
	input.ReplayedAppHash = hashStrings("different-app-hash")
	input.InvariantChecks = input.InvariantChecks[:3]
	input.PrefixMigrations[0].Safe = false

	report := BuildMigrationPhase0Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "single_chain_state_not_reproducible")
	require.Contains(t, report.Failed, "module_invariant_coverage")
	require.Contains(t, report.Failed, "prefix_migration:bank")
	require.NoError(t, report.Validate())
}

func TestMigrationPhase0ReadinessRequiresDeterministicGenesisImport(t *testing.T) {
	input := validMigrationPhase0Input()
	input.GenesisImports[0].Deterministic = false

	report := BuildMigrationPhase0Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "genesis_import:bank")
}

func TestMigrationPhase1ReadinessPassesCoreCommitments(t *testing.T) {
	report := BuildMigrationPhase1Readiness(validMigrationPhase1Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
}

func TestMigrationPhase1ReadinessFailsWhenNotSingleZoneOrCoreRootMissing(t *testing.T) {
	input := validMigrationPhase1Input()
	input.ZoneCount = 2
	input.AppHashIncludesCoreRoot = false
	input.MessageRoot = hashStrings("non-empty-message-root")

	report := BuildMigrationPhase1Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "app_hash_missing_core_root")
	require.Contains(t, report.Failed, "default_zone_count")
	require.Contains(t, report.Failed, "message_root_not_empty_queue")
	require.NoError(t, report.Validate())
}

func TestMigrationPhase1ReadinessRequiresRootQueriesAndProofMetadata(t *testing.T) {
	input := validMigrationPhase1Input()
	input.RootQueryAPIs = input.RootQueryAPIs[:1]
	input.ProofMetadata = input.ProofMetadata[:1]

	report := BuildMigrationPhase1Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "proof_registry_metadata")
	require.Contains(t, report.Failed, "root_query_apis")
}

func validMigrationPhase0Input() MigrationPhase0Input {
	appHash := hashStrings("single-chain-app-hash")
	return MigrationPhase0Input{
		ModuleBoundaryDocHash:     hashStrings("module-boundaries"),
		StateExportValidationHash: hashStrings("state-export-validation"),
		ExportedAppHash:           appHash,
		ReplayedAppHash:           appHash,
		GenesisImports: []GenesisImportCheck{
			genesisImport("bank"),
			genesisImport("staking"),
			genesisImport("slashing"),
			genesisImport("distribution"),
			genesisImport("aethercore"),
		},
		DynamicFeeBoundsTestHash: hashStrings("dynamic-fee-bounds"),
		InvariantChecks: []ModuleInvariantCheck{
			invariantCheck("bank", "supply"),
			invariantCheck("staking", "delegations"),
			invariantCheck("slashing", "tombstones"),
			invariantCheck("distribution", "outstanding_rewards"),
		},
		StoreV2CompatibilityHash: hashStrings("storev2-compatibility"),
		PrefixMigrations: []StatePrefixMigrationCheck{
			prefixMigration("bank", "bank/v1", "bank/v2"),
			prefixMigration("staking", "staking/v1", "staking/v2"),
		},
	}
}

func validMigrationPhase1Input() MigrationPhase1Input {
	emptyQueueRoot := hashStrings("empty-message-queues")
	return MigrationPhase1Input{
		AetherCoreModuleHash:    hashStrings("x-aethercore"),
		ZoneRegistryRoot:        hashStrings("default-zone-registry"),
		ZoneCount:               1,
		DefaultZoneID:           "default",
		DefaultZoneStateRoot:    hashStrings("default-zone-state-root"),
		MessageRoot:             emptyQueueRoot,
		EmptyQueueRoot:          emptyQueueRoot,
		ProofRegistryRoot:       hashStrings("proof-registry-root"),
		AppHashIncludesCoreRoot: true,
		CoreRootHash:            hashStrings("core-root"),
		RootQueryAPIs: []RootQueryAPICheck{
			rootQueryAPI("QueryZoneRoot", ProofRootZone),
			rootQueryAPI("QueryMessageRoot", ProofRootMessage),
			rootQueryAPI("QueryStorageRoot", ProofRootStorage),
		},
		ProofMetadata: []ProofRootMetadataCheck{
			proofMetadata(ProofRootZone),
			proofMetadata(ProofRootMessage),
			proofMetadata(ProofRootStorage),
		},
	}
}

func genesisImport(module string) GenesisImportCheck {
	root := hashStrings("genesis", module)
	return GenesisImportCheck{ModuleName: module, Active: true, Deterministic: true, ExportHash: root, ImportHash: root}
}

func invariantCheck(module, name string) ModuleInvariantCheck {
	return ModuleInvariantCheck{ModuleName: module, InvariantName: name, Covered: true, Deterministic: true, EvidenceHash: hashStrings("invariant", module, name)}
}

func prefixMigration(module, oldPrefix, newPrefix string) StatePrefixMigrationCheck {
	return StatePrefixMigrationCheck{
		ModuleName:      module,
		OldPrefix:       oldPrefix,
		NewPrefix:       newPrefix,
		MigrationHash:   hashStrings("migration", module, oldPrefix, newPrefix),
		ReversibleProof: hashStrings("migration-proof", module),
		Safe:            true,
	}
}

func rootQueryAPI(name string, rootType ProofRootType) RootQueryAPICheck {
	return RootQueryAPICheck{QueryName: name, RootType: rootType, Available: true, ResponseHash: hashStrings("root-query", name)}
}

func proofMetadata(rootType ProofRootType) ProofRootMetadataCheck {
	return ProofRootMetadataCheck{RootType: rootType, Height: 1, RootHash: hashStrings("root", string(rootType)), MetadataHash: hashStrings("metadata", string(rootType))}
}
