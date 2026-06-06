package types

import "testing"

func TestMigrationPathSpecCoversPhaseZeroThroughThree(t *testing.T) {
	spec, err := DefaultMigrationPathSpec()
	if err != nil {
		t.Fatalf("default migration path spec: %v", err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate migration path spec: %v", err)
	}
	if err := ValidateMigrationPathCoverage(); err != nil {
		t.Fatalf("migration path coverage: %v", err)
	}
	if len(spec.Phases) != 4 {
		t.Fatalf("expected 4 phases, got %d", len(spec.Phases))
	}

	phaseByID := map[MigrationPhaseID]MigrationPhase{}
	for _, phase := range spec.Phases {
		phaseByID[phase.PhaseID] = phase
	}

	phase0 := phaseByID[MigrationPhaseBaselineHardening]
	if len(phase0.Tasks) != 6 {
		t.Fatalf("expected phase 0 to cover 6 tasks, got %d", len(phase0.Tasks))
	}
	if len(phase0.ExitCriteria) != 3 {
		t.Fatalf("expected phase 0 to cover 3 exit criteria, got %d", len(phase0.ExitCriteria))
	}

	phase1 := phaseByID[MigrationPhaseCoreCommitments]
	if len(phase1.Tasks) != 6 {
		t.Fatalf("expected phase 1 to cover 6 tasks, got %d", len(phase1.Tasks))
	}
	if len(phase1.ExitCriteria) != 3 {
		t.Fatalf("expected phase 1 to cover 3 exit criteria, got %d", len(phase1.ExitCriteria))
	}

	phase2 := phaseByID[MigrationPhaseMessageBus]
	if len(phase2.Tasks) != 6 {
		t.Fatalf("expected phase 2 to cover 6 tasks, got %d", len(phase2.Tasks))
	}
	if len(phase2.ExitCriteria) != 3 {
		t.Fatalf("expected phase 2 to cover 3 exit criteria, got %d", len(phase2.ExitCriteria))
	}

	phase3 := phaseByID[MigrationPhaseZoneExtraction]
	if len(phase3.Tasks) != 6 {
		t.Fatalf("expected phase 3 to cover 6 tasks, got %d", len(phase3.Tasks))
	}
	if len(phase3.ExitCriteria) != 3 {
		t.Fatalf("expected phase 3 to cover 3 exit criteria, got %d", len(phase3.ExitCriteria))
	}
}

func TestMigrationPathSpecRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultMigrationPathSpec()
	if err != nil {
		t.Fatalf("default migration path spec: %v", err)
	}

	reordered, err := BuildMigrationPathSpec([]MigrationPhase{
		migrationPhase(MigrationPhaseZoneExtraction, "Phase 3: Zone Extraction", MigrationPhase3Tasks(), MigrationPhase3ExitCriteria()),
		migrationPhase(MigrationPhaseBaselineHardening, "Phase 0: Baseline Hardening", MigrationPhase0Tasks(), MigrationPhase0ExitCriteria()),
		migrationPhase(MigrationPhaseMessageBus, "Phase 2: Message Bus", MigrationPhase2Tasks(), MigrationPhase2ExitCriteria()),
		migrationPhase(MigrationPhaseCoreCommitments, "Phase 1: Core Commitments", MigrationPhase1Tasks(), MigrationPhase1ExitCriteria()),
	})
	if err != nil {
		t.Fatalf("build reordered migration path spec: %v", err)
	}
	if reordered.Root != defaultSpec.Root {
		t.Fatalf("canonical migration root mismatch: %s != %s", reordered.Root, defaultSpec.Root)
	}

	if _, err := BuildMigrationPathSpec([]MigrationPhase{defaultSpec.Phases[0], defaultSpec.Phases[0]}); err == nil {
		t.Fatal("expected duplicate migration phases to fail")
	}

	tampered := defaultSpec
	tampered.Phases[0].Tasks[0].DescriptorHash = hashParts("tampered migration task")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered migration task hash to fail")
	}
}

func TestBaselineHardeningEvidenceRequiresAllExitCriteria(t *testing.T) {
	evidence := validBaselineHardeningEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("baseline evidence should validate: %v", err)
	}

	notExportable := evidence
	notExportable.StateExportable = false
	notExportable.EvidenceHash = ComputeBaselineHardeningEvidenceHash(notExportable)
	if err := notExportable.Validate(); err == nil {
		t.Fatal("expected non-exportable baseline evidence to fail")
	}

	missingInvariantCoverage := evidence
	missingInvariantCoverage.InvariantCoverage = false
	missingInvariantCoverage.EvidenceHash = ComputeBaselineHardeningEvidenceHash(missingInvariantCoverage)
	if err := missingInvariantCoverage.Validate(); err == nil {
		t.Fatal("expected missing invariant coverage to fail")
	}

	unsafePrefixMigration := evidence
	unsafePrefixMigration.PrefixMigrationSafe = false
	unsafePrefixMigration.EvidenceHash = ComputeBaselineHardeningEvidenceHash(unsafePrefixMigration)
	if err := unsafePrefixMigration.Validate(); err == nil {
		t.Fatal("expected unsafe prefix migration to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different baseline evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered baseline evidence hash to fail")
	}
}

func TestCoreCommitmentMigrationEvidenceRequiresDefaultZoneRootsAndProofRegistry(t *testing.T) {
	evidence := validCoreCommitmentMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("core commitment evidence should validate: %v", err)
	}

	nonEmptyMessageRoot := evidence
	nonEmptyMessageRoot.EmptyMessageRoot = hashParts("not empty")
	nonEmptyMessageRoot.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(nonEmptyMessageRoot)
	if err := nonEmptyMessageRoot.Validate(); err == nil {
		t.Fatal("expected non-empty message root evidence to fail")
	}

	notSingleZone := evidence
	notSingleZone.SingleZoneMode = false
	notSingleZone.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(notSingleZone)
	if err := notSingleZone.Validate(); err == nil {
		t.Fatal("expected non single-zone evidence to fail")
	}

	missingMetadata := evidence
	missingMetadata.ProofRegistryMetadata = false
	missingMetadata.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(missingMetadata)
	if err := missingMetadata.Validate(); err == nil {
		t.Fatal("expected missing proof registry metadata to fail")
	}

	badZone := evidence
	badZone.DefaultZoneID = ZoneID("default")
	badZone.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(badZone)
	if err := badZone.Validate(); err == nil {
		t.Fatal("expected invalid default zone id to fail")
	}
}

func TestMessageBusMigrationEvidenceRequiresCommittedMessagesAndProofReceipts(t *testing.T) {
	evidence := validMessageBusMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("message bus evidence should validate: %v", err)
	}

	notCommitted := evidence
	notCommitted.MessagesCommitted = false
	notCommitted.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notCommitted)
	if err := notCommitted.Validate(); err == nil {
		t.Fatal("expected uncommitted message evidence to fail")
	}

	notDeterministic := evidence
	notDeterministic.LocalAsyncDeterministic = false
	notDeterministic.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notDeterministic)
	if err := notDeterministic.Validate(); err == nil {
		t.Fatal("expected non-deterministic local async evidence to fail")
	}

	notQueryable := evidence
	notQueryable.ReceiptProofQueryable = false
	notQueryable.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notQueryable)
	if err := notQueryable.Validate(); err == nil {
		t.Fatal("expected non-queryable receipt evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different message bus evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered message bus evidence hash to fail")
	}
}

func TestZoneExtractionMigrationEvidenceRequiresZoneIsolationAndPerBlockRoots(t *testing.T) {
	evidence := validZoneExtractionMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("zone extraction evidence should validate: %v", err)
	}

	financialNotRouted := evidence
	financialNotRouted.FinancialModulesRouted = false
	financialNotRouted.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(financialNotRouted)
	if err := financialNotRouted.Validate(); err == nil {
		t.Fatal("expected financial modules not routed evidence to fail")
	}

	identityNotIsolated := evidence
	identityNotIsolated.IdentityIsolated = false
	identityNotIsolated.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(identityNotIsolated)
	if err := identityNotIsolated.Validate(); err == nil {
		t.Fatal("expected non-isolated identity evidence to fail")
	}

	missingRoots := evidence
	missingRoots.ZoneRootsCommittedPerBlock = false
	missingRoots.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(missingRoots)
	if err := missingRoots.Validate(); err == nil {
		t.Fatal("expected missing per-block zone roots evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different zone extraction evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered zone extraction evidence hash to fail")
	}
}

func validBaselineHardeningEvidence() BaselineHardeningEvidence {
	evidence := BaselineHardeningEvidence{
		ModuleBoundaryDocsRoot:   hashParts("module boundary docs"),
		StateExportManifestHash:  hashParts("export manifest"),
		GenesisImportHash:        hashParts("genesis import"),
		DynamicFeeBoundsTestHash: hashParts("dynamic fee bounds tests"),
		LegacyInvariantRoot:      hashParts("staking slashing bank distribution invariants"),
		StoreV2AuditHash:         hashParts("store v2 compatibility audit"),
		UpgradeHandlerPrefixHash: hashParts("upgrade handler prefix migration"),
		StateReproducible:        true,
		StateExportable:          true,
		InvariantCoverage:        true,
		PrefixMigrationSafe:      true,
	}
	evidence.EvidenceHash = ComputeBaselineHardeningEvidenceHash(evidence)
	return evidence
}

func validCoreCommitmentMigrationEvidence() CoreCommitmentMigrationEvidence {
	evidence := CoreCommitmentMigrationEvidence{
		AetherCoreModuleHash:      hashParts("x/aethercore"),
		DefaultZoneDescriptorHash: hashParts("default zone descriptor"),
		DefaultZoneStateRoot:      hashParts("default zone state root"),
		EmptyMessageRoot:          EmptyRootHash,
		ProofRegistryRoot:         hashParts("proof root registry"),
		RootQueryAPIHash:          hashParts("root query APIs"),
		AppHashCoreRoot:           hashParts("app hash includes core root"),
		DefaultZoneID:             ZoneID("DEFAULT"),
		SingleZoneMode:            true,
		ProofRegistryMetadata:     true,
	}
	evidence.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(evidence)
	return evidence
}

func validMessageBusMigrationEvidence() MessageBusMigrationEvidence {
	evidence := MessageBusMigrationEvidence{
		MsgbusModuleHash:        hashParts("x/msgbus"),
		MessageCodecHash:        hashParts("canonical message encoding"),
		MessageIDDerivationHash: hashParts("message id derivation"),
		InboxStoreRoot:          hashParts("inbox store root"),
		OutboxStoreRoot:         hashParts("outbox store root"),
		ReceiptStoreRoot:        hashParts("receipt store root"),
		LocalExecutionRoot:      hashParts("local zone message execution"),
		ExpiryBounceRoot:        hashParts("expiry and bounce logic"),
		InclusionProofRoot:      hashParts("message inclusion proof"),
		MessagesCommitted:       true,
		LocalAsyncDeterministic: true,
		ReceiptProofQueryable:   true,
	}
	evidence.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(evidence)
	return evidence
}

func validZoneExtractionMigrationEvidence() ZoneExtractionMigrationEvidence {
	evidence := ZoneExtractionMigrationEvidence{
		FinancialZoneRoot:          hashParts("financial zone root"),
		IdentityZoneRoot:           hashParts("identity zone root"),
		ApplicationZoneRoot:        hashParts("application zone root"),
		ZoneKeeperRoot:             hashParts("zone keeper root"),
		ZonePrefixRoot:             hashParts("zone prefix root"),
		ZoneFeePolicyRoot:          hashParts("zone fee policy root"),
		ZoneExecutionSummaryRoot:   hashParts("zone execution summary root"),
		FinancialModulesRouted:     true,
		IdentityIsolated:           true,
		ZoneRootsCommittedPerBlock: true,
	}
	evidence.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(evidence)
	return evidence
}
