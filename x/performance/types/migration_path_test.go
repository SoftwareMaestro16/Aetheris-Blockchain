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

func TestMigrationPhase2ReadinessPassesMessageBus(t *testing.T) {
	report := BuildMigrationPhase2Readiness(validMigrationPhase2Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
}

func TestMigrationPhase2ReadinessFailsMissingStoreAndNondeterministicLocalExecution(t *testing.T) {
	input := validMigrationPhase2Input()
	input.Stores = input.Stores[:2]
	input.LocalExecution.Deterministic = false
	input.Safety.InclusionProofRoot = ""

	report := BuildMigrationPhase2Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "expiry_bounce_inclusion_receipt_proofs")
	require.Contains(t, report.Failed, "inbox_outbox_receipt_stores")
	require.Contains(t, report.Failed, "local_zone_message_execution")
	require.NoError(t, report.Validate())
}

func TestMigrationPhase3ReadinessPassesZoneExtraction(t *testing.T) {
	report := BuildMigrationPhase3Readiness(validMigrationPhase3Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
}

func TestMigrationPhase3ReadinessFailsIncompleteExtractionAndUncommittedRoots(t *testing.T) {
	input := validMigrationPhase3Input()
	input.FinancialZone.Extracted = false
	input.BankFeesTokenfactoryDEXInFinancial = false
	input.IdentityIsolatedActivation = false
	input.ZoneRootsCommittedPerBlock = false

	report := BuildMigrationPhase3Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "financial_zone_modules")
	require.Contains(t, report.Failed, "identity_zone_isolated_activation")
	require.Contains(t, report.Failed, "zone_extraction:financial")
	require.Contains(t, report.Failed, "zone_roots_committed_per_block")
	require.NoError(t, report.Validate())
}

func TestMigrationPhase4ReadinessPassesShardingRuntime(t *testing.T) {
	report := BuildMigrationPhase4Readiness(validMigrationPhase4Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
}

func TestMigrationPhase4ReadinessFailsSingleShardNondeterministicSchedulerAndLostMessages(t *testing.T) {
	input := validMigrationPhase4Input()
	input.ShardDescriptors = input.ShardDescriptors[:1]
	input.SplitMergeScheduler.Deterministic = false
	input.Migration.SurvivesLayoutChange = false
	input.InFlightMessagesSurviveChange = false

	report := BuildMigrationPhase4Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "zones_support_multiple_shards")
	require.Contains(t, report.Failed, "split_merge_scheduler")
	require.Contains(t, report.Failed, "deterministic_shard_migration")
	require.Contains(t, report.Failed, "in_flight_messages_survive_layout_changes")
	require.NoError(t, report.Validate())
}

func TestMigrationPhase4ReadinessRequiresPerShardInboxOutboxAndRouteKeys(t *testing.T) {
	input := validMigrationPhase4Input()
	input.ShardDescriptors[0].RouteKeyRoot = ""
	input.ShardDescriptors[1].InboxRoot = ""

	report := BuildMigrationPhase4Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "per_shard_runtime_descriptors")
}

func TestMigrationPhase5ReadinessPassesAVM20(t *testing.T) {
	report := BuildMigrationPhase5Readiness(validMigrationPhase5Input())
	require.True(t, report.Passed, report.Failed)
	require.Empty(t, report.Failed)
	require.NoError(t, report.Validate())
}

func TestMigrationPhase5ReadinessFailsMissingInterpreterGasProofSyscallAndProofRoot(t *testing.T) {
	input := validMigrationPhase5Input()
	input.Interpreter.Implemented = false
	input.GasTable.Deterministic = false
	input.ProofVerificationSyscalls[0].Metered = false
	input.ContractStateProofRoot = ""
	input.ContractZoneDeterministic = false

	report := BuildMigrationPhase5Readiness(input)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "avm_component:interpreter")
	require.Contains(t, report.Failed, "avm_component:gas_table")
	require.Contains(t, report.Failed, "proof_verification_syscalls")
	require.Contains(t, report.Failed, "contract_state_proof_root")
	require.Contains(t, report.Failed, "contract_zone_deterministic_execution")
	require.NoError(t, report.Validate())
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

func validMigrationPhase2Input() MigrationPhase2Input {
	return MigrationPhase2Input{
		MsgBusModuleHash: hashStrings("x-msgbus"),
		Encoding: MsgBusEncodingCheck{
			CodecHash:        hashStrings("msgbus-codec"),
			MessageIDRoot:    hashStrings("msgbus-message-ids"),
			DeterministicIDs: true,
		},
		Stores: []MsgBusStoreCheck{
			msgBusStore("inbox"),
			msgBusStore("outbox"),
			msgBusStore("receipt"),
		},
		LocalExecution: MsgBusExecutionCheck{
			ExecutionRoot:   hashStrings("msgbus-local-execution"),
			Deterministic:   true,
			ExecutedLocally: true,
		},
		Safety: MsgBusSafetyCheck{
			ExpiryRoot:         hashStrings("msgbus-expiry"),
			BounceRoot:         hashStrings("msgbus-bounce"),
			InclusionProofRoot: hashStrings("msgbus-inclusion-proofs"),
			ReceiptsProofRoot:  hashStrings("msgbus-receipt-proofs"),
		},
		FirstClassObjectRoot: hashStrings("first-class-message-objects"),
	}
}

func validMigrationPhase3Input() MigrationPhase3Input {
	return MigrationPhase3Input{
		FinancialZone:                      zoneExtraction("financial", []string{"bank", "fees", "tokenfactory", "dex"}),
		IdentityZone:                       zoneExtraction("identity", []string{"identity"}),
		ApplicationZone:                    zoneExtraction("application", []string{"scheduler", "workflow"}),
		BankFeesTokenfactoryDEXInFinancial: true,
		IdentityIsolatedActivation:         true,
		ZoneRootsCommittedPerBlock:         true,
		ZoneCommitmentRoot:                 hashStrings("zone-commitments-per-block"),
	}
}

func validMigrationPhase4Input() MigrationPhase4Input {
	return MigrationPhase4Input{
		ShardsModuleHash:          hashStrings("x-shards"),
		ZoneID:                    "financial",
		ShardLayoutDescriptorRoot: hashStrings("shard-layout-descriptors"),
		RouteKeyCalculationHash:   hashStrings("route-key-calculation"),
		ShardDescriptors: []ShardRuntimeDescriptorCheck{
			shardRuntimeDescriptor("shard-0001"),
			shardRuntimeDescriptor("shard-0002"),
		},
		RootAggregationHash: hashStrings("shard-root-aggregation"),
		SplitMergeScheduler: ShardSplitMergeSchedulerCheck{
			SchedulerRoot:     hashStrings("shard-scheduler"),
			SplitDecisionRoot: hashStrings("shard-split-decisions"),
			MergeDecisionRoot: hashStrings("shard-merge-decisions"),
			Deterministic:     true,
		},
		Migration: ShardMigrationCheck{
			MigrationRoot:          hashStrings("shard-migration"),
			OldLayoutHash:          hashStrings("old-shard-layout"),
			NewLayoutHash:          hashStrings("new-shard-layout"),
			InFlightMessageRoot:    hashStrings("in-flight-message-root"),
			SurvivesLayoutChange:   true,
			DeterministicMigration: true,
		},
		ZonesSupportMultipleShards:    true,
		IndependentWorkloadsParallel:  true,
		InFlightMessagesSurviveChange: true,
	}
}

func validMigrationPhase5Input() MigrationPhase5Input {
	return MigrationPhase5Input{
		BytecodeFormat:         avm20Component("bytecode_format"),
		Interpreter:            avm20Component("interpreter"),
		GasTable:               avm20Component("gas_table"),
		ContractStorageAdapter: avm20Component("contract_storage_adapter"),
		ABIRegistry:            avm20Component("abi_registry"),
		MessageSyscalls: []AVM20SyscallCheck{
			avm20Syscall("emit_message"),
			avm20Syscall("resolve_promise"),
		},
		ProofVerificationSyscalls: []AVM20SyscallCheck{
			avm20Syscall("verify_account_proof"),
			avm20Syscall("verify_contract_storage_proof"),
		},
		ContractZoneDeterministic: true,
		AsyncMessageEmissionRoot:  hashStrings("avm-async-message-emission"),
		ContractStateProofRoot:    hashStrings("avm-contract-state-proofs"),
		ContractZoneExecutionRoot: hashStrings("avm-contract-zone-execution"),
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

func msgBusStore(name string) MsgBusStoreCheck {
	return MsgBusStoreCheck{StoreName: name, RootHash: hashStrings("msgbus-store", name), Committed: true}
}

func zoneExtraction(zoneID string, modules []string) ZoneExtractionCheck {
	return ZoneExtractionCheck{
		ZoneID:               zoneID,
		Extracted:            true,
		KeeperHash:           hashStrings("zone-keeper", zoneID),
		StatePrefixRoot:      hashStrings("zone-prefix", zoneID),
		FeePolicyHash:        hashStrings("zone-fee-policy", zoneID),
		ExecutionSummaryHash: hashStrings("zone-summary", zoneID),
		CommittedRoot:        hashStrings("zone-root", zoneID),
		Modules:              modules,
	}
}

func shardRuntimeDescriptor(shardID string) ShardRuntimeDescriptorCheck {
	return ShardRuntimeDescriptorCheck{
		ShardID:           shardID,
		LayoutHash:        hashStrings("shard-layout", shardID),
		RouteKeyRoot:      hashStrings("shard-route-key", shardID),
		InboxRoot:         hashStrings("shard-inbox", shardID),
		OutboxRoot:        hashStrings("shard-outbox", shardID),
		ShardRoot:         hashStrings("shard-state-root", shardID),
		ParallelGroupHash: hashStrings("shard-parallel-group", shardID),
		Active:            true,
	}
}

func avm20Component(name string) AVM20ComponentCheck {
	return AVM20ComponentCheck{
		ComponentName: name,
		ComponentHash: hashStrings("avm20-component", name),
		Implemented:   true,
		Deterministic: true,
	}
}

func avm20Syscall(name string) AVM20SyscallCheck {
	return AVM20SyscallCheck{
		SyscallName: name,
		SyscallHash: hashStrings("avm20-syscall", name),
		Metered:     true,
		Enabled:     true,
	}
}
