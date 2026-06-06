package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type MigrationPhase string

const (
	MigrationPhase0BaselineHardening MigrationPhase = "phase_0_baseline_hardening"
	MigrationPhase1CoreCommitments   MigrationPhase = "phase_1_core_commitments"
	MigrationPhase2MessageBus        MigrationPhase = "phase_2_message_bus"
	MigrationPhase3ZoneExtraction    MigrationPhase = "phase_3_zone_extraction"
)

type GenesisImportCheck struct {
	ModuleName    string
	Active        bool
	Deterministic bool
	ExportHash    string
	ImportHash    string
}

type ModuleInvariantCheck struct {
	ModuleName    string
	InvariantName string
	Covered       bool
	Deterministic bool
	EvidenceHash  string
}

type StatePrefixMigrationCheck struct {
	ModuleName      string
	OldPrefix       string
	NewPrefix       string
	MigrationHash   string
	ReversibleProof string
	Safe            bool
}

type MigrationPhase0Input struct {
	ModuleBoundaryDocHash     string
	StateExportValidationHash string
	ExportedAppHash           string
	ReplayedAppHash           string
	GenesisImports            []GenesisImportCheck
	DynamicFeeBoundsTestHash  string
	InvariantChecks           []ModuleInvariantCheck
	StoreV2CompatibilityHash  string
	PrefixMigrations          []StatePrefixMigrationCheck
}

type RootQueryAPICheck struct {
	QueryName    string
	RootType     ProofRootType
	Available    bool
	ResponseHash string
}

type ProofRootMetadataCheck struct {
	RootType     ProofRootType
	Height       uint64
	RootHash     string
	MetadataHash string
}

type MigrationPhase1Input struct {
	AetherCoreModuleHash    string
	ZoneRegistryRoot        string
	ZoneCount               uint32
	DefaultZoneID           string
	DefaultZoneStateRoot    string
	MessageRoot             string
	EmptyQueueRoot          string
	ProofRegistryRoot       string
	RootQueryAPIs           []RootQueryAPICheck
	ProofMetadata           []ProofRootMetadataCheck
	AppHashIncludesCoreRoot bool
	CoreRootHash            string
}

type MsgBusStoreCheck struct {
	StoreName string
	RootHash  string
	Committed bool
}

type MsgBusEncodingCheck struct {
	CodecHash        string
	MessageIDRoot    string
	DeterministicIDs bool
}

type MsgBusExecutionCheck struct {
	ExecutionRoot   string
	Deterministic   bool
	ExecutedLocally bool
}

type MsgBusSafetyCheck struct {
	ExpiryRoot         string
	BounceRoot         string
	InclusionProofRoot string
	ReceiptsProofRoot  string
}

type MigrationPhase2Input struct {
	MsgBusModuleHash     string
	Encoding             MsgBusEncodingCheck
	Stores               []MsgBusStoreCheck
	LocalExecution       MsgBusExecutionCheck
	Safety               MsgBusSafetyCheck
	FirstClassObjectRoot string
}

type ZoneExtractionCheck struct {
	ZoneID               string
	Extracted            bool
	KeeperHash           string
	StatePrefixRoot      string
	FeePolicyHash        string
	ExecutionSummaryHash string
	CommittedRoot        string
	Modules              []string
}

type MigrationPhase3Input struct {
	FinancialZone                      ZoneExtractionCheck
	IdentityZone                       ZoneExtractionCheck
	ApplicationZone                    ZoneExtractionCheck
	BankFeesTokenfactoryDEXInFinancial bool
	IdentityIsolatedActivation         bool
	ZoneRootsCommittedPerBlock         bool
	ZoneCommitmentRoot                 string
}

type MigrationReadinessReport struct {
	Phase      MigrationPhase
	Passed     bool
	Failed     []string
	Evidence   []string
	ReportHash string
}

func BuildMigrationPhase0Readiness(input MigrationPhase0Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	if err := validateHexHash("migration phase 0 module boundary documentation hash", input.ModuleBoundaryDocHash); err != nil {
		failed = append(failed, "module_boundary_documentation")
	} else {
		evidence = append(evidence, "module_boundary_documentation:"+input.ModuleBoundaryDocHash)
	}
	if err := validateHexHash("migration phase 0 state export validation hash", input.StateExportValidationHash); err != nil {
		failed = append(failed, "state_export_validation")
	} else {
		evidence = append(evidence, "state_export_validation:"+input.StateExportValidationHash)
	}
	if input.ExportedAppHash == "" || input.ReplayedAppHash == "" || input.ExportedAppHash != input.ReplayedAppHash {
		failed = append(failed, "single_chain_state_not_reproducible")
	} else if err := validateHexHash("migration phase 0 exported app hash", input.ExportedAppHash); err != nil {
		failed = append(failed, "single_chain_state_hash_invalid")
	} else {
		evidence = append(evidence, "reproducible_export:"+input.ExportedAppHash)
	}
	if len(input.GenesisImports) == 0 {
		failed = append(failed, "active_module_genesis_imports_missing")
	}
	for _, check := range input.GenesisImports {
		if err := check.Validate(); err != nil {
			failed = append(failed, "genesis_import:"+check.ModuleName)
		} else if check.Active {
			evidence = append(evidence, "genesis_import:"+check.ModuleName+":"+check.ImportHash)
		}
	}
	if err := validateHexHash("migration phase 0 dynamic fee bounds test hash", input.DynamicFeeBoundsTestHash); err != nil {
		failed = append(failed, "dynamic_fee_bounds_tests")
	} else {
		evidence = append(evidence, "dynamic_fee_bounds_tests:"+input.DynamicFeeBoundsTestHash)
	}
	if err := validateRequiredInvariantCoverage(input.InvariantChecks); err != nil {
		failed = append(failed, "module_invariant_coverage")
	} else {
		evidence = append(evidence, "module_invariant_coverage:"+hashInvariantChecks(input.InvariantChecks))
	}
	if err := validateHexHash("migration phase 0 Store v2 compatibility audit hash", input.StoreV2CompatibilityHash); err != nil {
		failed = append(failed, "store_v2_compatibility_audit")
	} else {
		evidence = append(evidence, "store_v2_compatibility_audit:"+input.StoreV2CompatibilityHash)
	}
	if len(input.PrefixMigrations) == 0 {
		failed = append(failed, "upgrade_prefix_migrations_missing")
	}
	for _, migration := range input.PrefixMigrations {
		if err := migration.Validate(); err != nil {
			failed = append(failed, "prefix_migration:"+migration.ModuleName)
		} else {
			evidence = append(evidence, "prefix_migration:"+migration.ModuleName+":"+migration.MigrationHash)
		}
	}
	report := MigrationReadinessReport{
		Phase:    MigrationPhase0BaselineHardening,
		Passed:   len(failed) == 0,
		Failed:   normalizeStringSet(failed),
		Evidence: normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase1Readiness(input MigrationPhase1Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label string
		hash  string
	}{
		{"aethercore_module", input.AetherCoreModuleHash},
		{"zone_registry_root", input.ZoneRegistryRoot},
		{"default_zone_state_root", input.DefaultZoneStateRoot},
		{"message_root", input.MessageRoot},
		{"empty_queue_root", input.EmptyQueueRoot},
		{"proof_registry_root", input.ProofRegistryRoot},
		{"core_root_hash", input.CoreRootHash},
	} {
		if err := validateHexHash("migration phase 1 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if input.ZoneCount != 1 {
		failed = append(failed, "default_zone_count")
	}
	if strings.TrimSpace(input.DefaultZoneID) == "" {
		failed = append(failed, "default_zone_id")
	}
	if input.MessageRoot != input.EmptyQueueRoot {
		failed = append(failed, "message_root_not_empty_queue")
	}
	if !input.AppHashIncludesCoreRoot {
		failed = append(failed, "app_hash_missing_core_root")
	}
	if err := validateRootQueryAPIs(input.RootQueryAPIs); err != nil {
		failed = append(failed, "root_query_apis")
	} else {
		evidence = append(evidence, "root_query_apis:"+hashRootQueryAPIs(input.RootQueryAPIs))
	}
	if err := validateProofRootMetadata(input.ProofMetadata); err != nil {
		failed = append(failed, "proof_registry_metadata")
	} else {
		evidence = append(evidence, "proof_registry_metadata:"+hashProofMetadata(input.ProofMetadata))
	}
	report := MigrationReadinessReport{
		Phase:    MigrationPhase1CoreCommitments,
		Passed:   len(failed) == 0,
		Failed:   normalizeStringSet(failed),
		Evidence: normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase2Readiness(input MigrationPhase2Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label string
		hash  string
	}{
		{"msgbus_module", input.MsgBusModuleHash},
		{"first_class_message_objects", input.FirstClassObjectRoot},
	} {
		if err := validateHexHash("migration phase 2 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if err := input.Encoding.Validate(); err != nil {
		failed = append(failed, "message_encoding_and_ids")
	} else {
		evidence = append(evidence, "message_encoding_and_ids:"+input.Encoding.MessageIDRoot)
	}
	if err := validateMsgBusStores(input.Stores); err != nil {
		failed = append(failed, "inbox_outbox_receipt_stores")
	} else {
		evidence = append(evidence, "message_stores:"+hashMsgBusStores(input.Stores))
	}
	if err := input.LocalExecution.Validate(); err != nil {
		failed = append(failed, "local_zone_message_execution")
	} else {
		evidence = append(evidence, "local_zone_message_execution:"+input.LocalExecution.ExecutionRoot)
	}
	if err := input.Safety.Validate(); err != nil {
		failed = append(failed, "expiry_bounce_inclusion_receipt_proofs")
	} else {
		evidence = append(evidence, "message_safety:"+hashMsgBusSafety(input.Safety))
	}
	report := MigrationReadinessReport{
		Phase:    MigrationPhase2MessageBus,
		Passed:   len(failed) == 0,
		Failed:   normalizeStringSet(failed),
		Evidence: normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase3Readiness(input MigrationPhase3Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, zone := range []ZoneExtractionCheck{input.FinancialZone, input.IdentityZone, input.ApplicationZone} {
		if err := zone.Validate(); err != nil {
			failed = append(failed, "zone_extraction:"+zone.ZoneID)
		} else {
			evidence = append(evidence, "zone_extraction:"+zone.ZoneID+":"+zone.CommittedRoot)
		}
	}
	if !input.BankFeesTokenfactoryDEXInFinancial {
		failed = append(failed, "financial_zone_modules")
	}
	if !input.IdentityIsolatedActivation {
		failed = append(failed, "identity_zone_isolated_activation")
	}
	if !input.ZoneRootsCommittedPerBlock {
		failed = append(failed, "zone_roots_committed_per_block")
	}
	if err := validateHexHash("migration phase 3 zone commitment root", input.ZoneCommitmentRoot); err != nil {
		failed = append(failed, "zone_commitment_root")
	} else {
		evidence = append(evidence, "zone_commitment_root:"+input.ZoneCommitmentRoot)
	}
	report := MigrationReadinessReport{
		Phase:    MigrationPhase3ZoneExtraction,
		Passed:   len(failed) == 0,
		Failed:   normalizeStringSet(failed),
		Evidence: normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func (i MigrationPhase0Input) Normalize() MigrationPhase0Input {
	i.ModuleBoundaryDocHash = normalizeLowerHex(i.ModuleBoundaryDocHash)
	i.StateExportValidationHash = normalizeLowerHex(i.StateExportValidationHash)
	i.ExportedAppHash = normalizeLowerHex(i.ExportedAppHash)
	i.ReplayedAppHash = normalizeLowerHex(i.ReplayedAppHash)
	i.DynamicFeeBoundsTestHash = normalizeLowerHex(i.DynamicFeeBoundsTestHash)
	i.StoreV2CompatibilityHash = normalizeLowerHex(i.StoreV2CompatibilityHash)
	for idx := range i.GenesisImports {
		i.GenesisImports[idx] = i.GenesisImports[idx].Normalize()
	}
	sort.SliceStable(i.GenesisImports, func(left, right int) bool {
		return i.GenesisImports[left].ModuleName < i.GenesisImports[right].ModuleName
	})
	for idx := range i.InvariantChecks {
		i.InvariantChecks[idx] = i.InvariantChecks[idx].Normalize()
	}
	sort.SliceStable(i.InvariantChecks, func(left, right int) bool {
		return invariantKey(i.InvariantChecks[left]) < invariantKey(i.InvariantChecks[right])
	})
	for idx := range i.PrefixMigrations {
		i.PrefixMigrations[idx] = i.PrefixMigrations[idx].Normalize()
	}
	sort.SliceStable(i.PrefixMigrations, func(left, right int) bool {
		return i.PrefixMigrations[left].ModuleName < i.PrefixMigrations[right].ModuleName
	})
	return i
}

func (i MigrationPhase1Input) Normalize() MigrationPhase1Input {
	i.AetherCoreModuleHash = normalizeLowerHex(i.AetherCoreModuleHash)
	i.ZoneRegistryRoot = normalizeLowerHex(i.ZoneRegistryRoot)
	i.DefaultZoneID = strings.TrimSpace(i.DefaultZoneID)
	i.DefaultZoneStateRoot = normalizeLowerHex(i.DefaultZoneStateRoot)
	i.MessageRoot = normalizeLowerHex(i.MessageRoot)
	i.EmptyQueueRoot = normalizeLowerHex(i.EmptyQueueRoot)
	i.ProofRegistryRoot = normalizeLowerHex(i.ProofRegistryRoot)
	i.CoreRootHash = normalizeLowerHex(i.CoreRootHash)
	for idx := range i.RootQueryAPIs {
		i.RootQueryAPIs[idx] = i.RootQueryAPIs[idx].Normalize()
	}
	sort.SliceStable(i.RootQueryAPIs, func(left, right int) bool {
		return i.RootQueryAPIs[left].QueryName < i.RootQueryAPIs[right].QueryName
	})
	for idx := range i.ProofMetadata {
		i.ProofMetadata[idx] = i.ProofMetadata[idx].Normalize()
	}
	sort.SliceStable(i.ProofMetadata, func(left, right int) bool {
		return string(i.ProofMetadata[left].RootType) < string(i.ProofMetadata[right].RootType)
	})
	return i
}

func (i MigrationPhase2Input) Normalize() MigrationPhase2Input {
	i.MsgBusModuleHash = normalizeLowerHex(i.MsgBusModuleHash)
	i.Encoding = i.Encoding.Normalize()
	for idx := range i.Stores {
		i.Stores[idx] = i.Stores[idx].Normalize()
	}
	sort.SliceStable(i.Stores, func(left, right int) bool {
		return i.Stores[left].StoreName < i.Stores[right].StoreName
	})
	i.LocalExecution = i.LocalExecution.Normalize()
	i.Safety = i.Safety.Normalize()
	i.FirstClassObjectRoot = normalizeLowerHex(i.FirstClassObjectRoot)
	return i
}

func (i MigrationPhase3Input) Normalize() MigrationPhase3Input {
	i.FinancialZone = i.FinancialZone.Normalize()
	i.IdentityZone = i.IdentityZone.Normalize()
	i.ApplicationZone = i.ApplicationZone.Normalize()
	i.ZoneCommitmentRoot = normalizeLowerHex(i.ZoneCommitmentRoot)
	return i
}

func (c GenesisImportCheck) Normalize() GenesisImportCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.ExportHash = normalizeLowerHex(c.ExportHash)
	c.ImportHash = normalizeLowerHex(c.ImportHash)
	return c
}

func (c GenesisImportCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration genesis import module", check.ModuleName); err != nil {
		return err
	}
	if !check.Active {
		return nil
	}
	if !check.Deterministic {
		return errors.New("migration active module genesis import must be deterministic")
	}
	if err := validateHexHash("migration genesis export hash", check.ExportHash); err != nil {
		return err
	}
	if err := validateHexHash("migration genesis import hash", check.ImportHash); err != nil {
		return err
	}
	if check.ExportHash != check.ImportHash {
		return errors.New("migration genesis import hash must reproduce export hash")
	}
	return nil
}

func (c ModuleInvariantCheck) Normalize() ModuleInvariantCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.InvariantName = strings.TrimSpace(c.InvariantName)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	return c
}

func (c ModuleInvariantCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration invariant module", check.ModuleName); err != nil {
		return err
	}
	if err := validateExecutionToken("migration invariant name", check.InvariantName); err != nil {
		return err
	}
	if !check.Covered || !check.Deterministic {
		return errors.New("migration invariant must be covered and deterministic")
	}
	return validateHexHash("migration invariant evidence hash", check.EvidenceHash)
}

func (c StatePrefixMigrationCheck) Normalize() StatePrefixMigrationCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.OldPrefix = strings.TrimSpace(c.OldPrefix)
	c.NewPrefix = strings.TrimSpace(c.NewPrefix)
	c.MigrationHash = normalizeLowerHex(c.MigrationHash)
	c.ReversibleProof = normalizeLowerHex(c.ReversibleProof)
	return c
}

func (c StatePrefixMigrationCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration prefix module", check.ModuleName); err != nil {
		return err
	}
	if check.OldPrefix == "" || check.NewPrefix == "" || check.OldPrefix == check.NewPrefix {
		return errors.New("migration state prefixes must be non-empty and changed")
	}
	if !check.Safe {
		return errors.New("migration state prefix migration must be marked safe")
	}
	if err := validateHexHash("migration prefix migration hash", check.MigrationHash); err != nil {
		return err
	}
	return validateHexHash("migration prefix reversible proof", check.ReversibleProof)
}

func (c RootQueryAPICheck) Normalize() RootQueryAPICheck {
	c.QueryName = strings.TrimSpace(c.QueryName)
	c.ResponseHash = normalizeLowerHex(c.ResponseHash)
	return c
}

func (c RootQueryAPICheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration root query API", check.QueryName); err != nil {
		return err
	}
	if !check.Available {
		return errors.New("migration root query API must be available")
	}
	if !IsProofRootType(check.RootType) {
		return errors.New("migration root query API root type is unsupported")
	}
	return validateHexHash("migration root query response hash", check.ResponseHash)
}

func (c ProofRootMetadataCheck) Normalize() ProofRootMetadataCheck {
	c.RootHash = normalizeLowerHex(c.RootHash)
	c.MetadataHash = normalizeLowerHex(c.MetadataHash)
	return c
}

func (c ProofRootMetadataCheck) Validate() error {
	check := c.Normalize()
	if !IsProofRootType(check.RootType) {
		return errors.New("migration proof metadata root type is unsupported")
	}
	if check.Height == 0 {
		return errors.New("migration proof metadata height must be positive")
	}
	if err := validateHexHash("migration proof metadata root hash", check.RootHash); err != nil {
		return err
	}
	return validateHexHash("migration proof metadata hash", check.MetadataHash)
}

func (c MsgBusStoreCheck) Normalize() MsgBusStoreCheck {
	c.StoreName = strings.TrimSpace(c.StoreName)
	c.RootHash = normalizeLowerHex(c.RootHash)
	return c
}

func (c MsgBusStoreCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration msgbus store name", check.StoreName); err != nil {
		return err
	}
	if !check.Committed {
		return errors.New("migration msgbus store must be committed")
	}
	return validateHexHash("migration msgbus store root", check.RootHash)
}

func (c MsgBusEncodingCheck) Normalize() MsgBusEncodingCheck {
	c.CodecHash = normalizeLowerHex(c.CodecHash)
	c.MessageIDRoot = normalizeLowerHex(c.MessageIDRoot)
	return c
}

func (c MsgBusEncodingCheck) Validate() error {
	check := c.Normalize()
	if !check.DeterministicIDs {
		return errors.New("migration msgbus message ids must be deterministic")
	}
	if err := validateHexHash("migration msgbus codec hash", check.CodecHash); err != nil {
		return err
	}
	return validateHexHash("migration msgbus message id root", check.MessageIDRoot)
}

func (c MsgBusExecutionCheck) Normalize() MsgBusExecutionCheck {
	c.ExecutionRoot = normalizeLowerHex(c.ExecutionRoot)
	return c
}

func (c MsgBusExecutionCheck) Validate() error {
	check := c.Normalize()
	if !check.Deterministic || !check.ExecutedLocally {
		return errors.New("migration msgbus local async execution must be deterministic and local")
	}
	return validateHexHash("migration msgbus local execution root", check.ExecutionRoot)
}

func (c MsgBusSafetyCheck) Normalize() MsgBusSafetyCheck {
	c.ExpiryRoot = normalizeLowerHex(c.ExpiryRoot)
	c.BounceRoot = normalizeLowerHex(c.BounceRoot)
	c.InclusionProofRoot = normalizeLowerHex(c.InclusionProofRoot)
	c.ReceiptsProofRoot = normalizeLowerHex(c.ReceiptsProofRoot)
	return c
}

func (c MsgBusSafetyCheck) Validate() error {
	check := c.Normalize()
	for _, item := range []struct {
		name  string
		value string
	}{
		{"migration msgbus expiry root", check.ExpiryRoot},
		{"migration msgbus bounce root", check.BounceRoot},
		{"migration msgbus inclusion proof root", check.InclusionProofRoot},
		{"migration msgbus receipts proof root", check.ReceiptsProofRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (c ZoneExtractionCheck) Normalize() ZoneExtractionCheck {
	c.ZoneID = strings.TrimSpace(c.ZoneID)
	c.KeeperHash = normalizeLowerHex(c.KeeperHash)
	c.StatePrefixRoot = normalizeLowerHex(c.StatePrefixRoot)
	c.FeePolicyHash = normalizeLowerHex(c.FeePolicyHash)
	c.ExecutionSummaryHash = normalizeLowerHex(c.ExecutionSummaryHash)
	c.CommittedRoot = normalizeLowerHex(c.CommittedRoot)
	c.Modules = normalizeStringSet(c.Modules)
	return c
}

func (c ZoneExtractionCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration extracted zone id", check.ZoneID); err != nil {
		return err
	}
	if !check.Extracted {
		return errors.New("migration zone must be extracted")
	}
	if len(check.Modules) == 0 {
		return errors.New("migration extracted zone requires modules")
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{"migration zone keeper hash", check.KeeperHash},
		{"migration zone state prefix root", check.StatePrefixRoot},
		{"migration zone fee policy hash", check.FeePolicyHash},
		{"migration zone execution summary hash", check.ExecutionSummaryHash},
		{"migration zone committed root", check.CommittedRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (r MigrationReadinessReport) Validate() error {
	if r.Phase != MigrationPhase0BaselineHardening &&
		r.Phase != MigrationPhase1CoreCommitments &&
		r.Phase != MigrationPhase2MessageBus &&
		r.Phase != MigrationPhase3ZoneExtraction {
		return errors.New("migration readiness phase is unsupported")
	}
	if r.Passed && len(r.Failed) > 0 {
		return errors.New("migration readiness passed report must not include failures")
	}
	if len(r.Evidence) == 0 {
		return errors.New("migration readiness evidence is required")
	}
	if r.ReportHash != ComputeMigrationReadinessReportHash(r) {
		return errors.New("migration readiness report hash mismatch")
	}
	return nil
}

func validateMsgBusStores(checks []MsgBusStoreCheck) error {
	required := map[string]struct{}{
		"inbox":   {},
		"outbox":  {},
		"receipt": {},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.StoreName]; found {
			delete(required, check.StoreName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing msgbus stores: %v", sortedMapKeys(required))
	}
	return nil
}

func validateRequiredInvariantCoverage(checks []ModuleInvariantCheck) error {
	required := map[string]struct{}{
		"staking":      {},
		"slashing":     {},
		"bank":         {},
		"distribution": {},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.ModuleName]; found {
			delete(required, check.ModuleName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing required invariants: %v", sortedMapKeys(required))
	}
	return nil
}

func validateRootQueryAPIs(checks []RootQueryAPICheck) error {
	required := map[ProofRootType]struct{}{
		ProofRootZone:    {},
		ProofRootMessage: {},
		ProofRootStorage: {},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.RootType]; found {
			delete(required, check.RootType)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing root query APIs: %v", required)
	}
	return nil
}

func validateProofRootMetadata(checks []ProofRootMetadataCheck) error {
	required := map[ProofRootType]struct{}{
		ProofRootZone:    {},
		ProofRootMessage: {},
		ProofRootStorage: {},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.RootType]; found {
			delete(required, check.RootType)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing proof metadata: %v", required)
	}
	return nil
}

func ComputeMigrationReadinessReportHash(report MigrationReadinessReport) string {
	failed := normalizeStringSet(report.Failed)
	evidence := normalizeStringSet(report.Evidence)
	parts := []string{"migration-readiness-report", string(report.Phase), fmt.Sprintf("%t", report.Passed)}
	parts = append(parts, failed...)
	parts = append(parts, evidence...)
	return hashStrings(parts...)
}

func hashInvariantChecks(checks []ModuleInvariantCheck) string {
	parts := []string{"migration-invariant-checks"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.ModuleName, check.InvariantName, check.EvidenceHash)
	}
	return hashStrings(parts...)
}

func hashRootQueryAPIs(checks []RootQueryAPICheck) string {
	parts := []string{"migration-root-query-apis"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.QueryName, string(check.RootType), check.ResponseHash)
	}
	return hashStrings(parts...)
}

func hashProofMetadata(checks []ProofRootMetadataCheck) string {
	parts := []string{"migration-proof-metadata"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, string(check.RootType), fmt.Sprintf("%020d", check.Height), check.RootHash, check.MetadataHash)
	}
	return hashStrings(parts...)
}

func hashMsgBusStores(checks []MsgBusStoreCheck) string {
	parts := []string{"migration-msgbus-stores"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.StoreName, check.RootHash, fmt.Sprintf("%t", check.Committed))
	}
	return hashStrings(parts...)
}

func hashMsgBusSafety(check MsgBusSafetyCheck) string {
	check = check.Normalize()
	return hashStrings("migration-msgbus-safety", check.ExpiryRoot, check.BounceRoot, check.InclusionProofRoot, check.ReceiptsProofRoot)
}

func invariantKey(check ModuleInvariantCheck) string {
	return check.ModuleName + "/" + check.InvariantName
}

func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
