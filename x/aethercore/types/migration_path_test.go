package types

import "testing"

func TestMigrationPathSpecCoversPhaseZeroAndOne(t *testing.T) {
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
	if len(spec.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(spec.Phases))
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
}

func TestMigrationPathSpecRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultMigrationPathSpec()
	if err != nil {
		t.Fatalf("default migration path spec: %v", err)
	}

	reordered, err := BuildMigrationPathSpec([]MigrationPhase{
		migrationPhase(MigrationPhaseCoreCommitments, "Phase 1: Core Commitments", MigrationPhase1Tasks(), MigrationPhase1ExitCriteria()),
		migrationPhase(MigrationPhaseBaselineHardening, "Phase 0: Baseline Hardening", MigrationPhase0Tasks(), MigrationPhase0ExitCriteria()),
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
