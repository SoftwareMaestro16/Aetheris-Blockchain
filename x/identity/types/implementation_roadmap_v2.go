package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityImplementationRoadmapPhaseIDV2 string
type IdentityImplementationRoadmapTaskIDV2 string
type IdentityImplementationRoadmapExitIDV2 string

const (
	IdentityRoadmapPhase0SpecVectorsV2     IdentityImplementationRoadmapPhaseIDV2 = "phase_0_specification_and_test_vectors"
	IdentityRoadmapPhase1CoreActivationV2  IdentityImplementationRoadmapPhaseIDV2 = "phase_1_core_registry_activation"
	IdentityRoadmapPhase2UnifiedResolverV2 IdentityImplementationRoadmapPhaseIDV2 = "phase_2_unified_resolver"

	IdentityRoadmapTaskCanonicalNameNormalizationV2 IdentityImplementationRoadmapTaskIDV2 = "define_canonical_name_normalization"
	IdentityRoadmapTaskDomainProofHashFormatsV2     IdentityImplementationRoadmapTaskIDV2 = "define_domain_hash_and_proof_hash_formats"
	IdentityRoadmapTaskProtobufStateSchemasV2       IdentityImplementationRoadmapTaskIDV2 = "define_protobuf_state_schemas"
	IdentityRoadmapTaskStoreV2KeyLayoutV2           IdentityImplementationRoadmapTaskIDV2 = "define_store_v2_key_layout"
	IdentityRoadmapTaskGovernanceParamsV2           IdentityImplementationRoadmapTaskIDV2 = "define_governance_parameter_set"
	IdentityRoadmapTaskResolutionProofVectorsV2     IdentityImplementationRoadmapTaskIDV2 = "produce_resolution_proof_test_vectors"
	IdentityRoadmapTaskLifecycleVectorsV2           IdentityImplementationRoadmapTaskIDV2 = "produce_lifecycle_transition_test_vectors"

	IdentityRoadmapTaskIdentityCoreModuleV2 IdentityImplementationRoadmapTaskIDV2 = "implement_identity_core_module"
	IdentityRoadmapTaskCoreLifecycleV2      IdentityImplementationRoadmapTaskIDV2 = "implement_registration_renewal_transfer_expiry"
	IdentityRoadmapTaskNFTBindingV2         IdentityImplementationRoadmapTaskIDV2 = "implement_nft_binding"
	IdentityRoadmapTaskOwnerExpiryIndexesV2 IdentityImplementationRoadmapTaskIDV2 = "implement_owner_and_expiry_indexes"
	IdentityRoadmapTaskCoreQueriesV2        IdentityImplementationRoadmapTaskIDV2 = "implement_core_queries"
	IdentityRoadmapTaskInvariantChecksV2    IdentityImplementationRoadmapTaskIDV2 = "add_invariant_checks"

	IdentityRoadmapTaskResolverModuleV2       IdentityImplementationRoadmapTaskIDV2 = "implement_resolver_module"
	IdentityRoadmapTaskPrimaryResolutionV2    IdentityImplementationRoadmapTaskIDV2 = "implement_primary_address_resolution"
	IdentityRoadmapTaskContractTargetsV2      IdentityImplementationRoadmapTaskIDV2 = "implement_contract_targets"
	IdentityRoadmapTaskServiceEndpointsV2     IdentityImplementationRoadmapTaskIDV2 = "implement_service_endpoints"
	IdentityRoadmapTaskInterfaceDescriptorsV2 IdentityImplementationRoadmapTaskIDV2 = "implement_interface_descriptors"
	IdentityRoadmapTaskRoutingMetadataV2      IdentityImplementationRoadmapTaskIDV2 = "implement_routing_metadata"
	IdentityRoadmapTaskReverseResolutionV2    IdentityImplementationRoadmapTaskIDV2 = "implement_reverse_resolution"
	IdentityRoadmapTaskBatchResolverUpdatesV2 IdentityImplementationRoadmapTaskIDV2 = "implement_batch_resolver_updates"

	IdentityRoadmapExitSignableHashableVectorsV2 IdentityImplementationRoadmapExitIDV2 = "all_signable_and_hashable_identity_objects_have_test_vectors"
	IdentityRoadmapExitLifecycleDeterminismV2    IdentityImplementationRoadmapExitIDV2 = "all_lifecycle_states_have_deterministic_transition_tests"
	IdentityRoadmapExitStorePrefixesFinalizedV2  IdentityImplementationRoadmapExitIDV2 = "store_key_prefixes_are_finalized"

	IdentityRoadmapExitOnChainOwnershipV2     IdentityImplementationRoadmapExitIDV2 = "aet_domain_ownership_is_fully_on_chain"
	IdentityRoadmapExitAtomicNFTOwnershipV2   IdentityImplementationRoadmapExitIDV2 = "nft_and_registry_ownership_remain_atomic"
	IdentityRoadmapExitExportImportRegistryV2 IdentityImplementationRoadmapExitIDV2 = "export_import_preserves_registry_state"

	IdentityRoadmapExitUnifiedTargetsV2       IdentityImplementationRoadmapExitIDV2 = "unified_resolver_supports_wallet_contract_service_interface_routing_targets"
	IdentityRoadmapExitReverseConsistencyV2   IdentityImplementationRoadmapExitIDV2 = "reverse_resolution_verifies_forward_consistency"
	IdentityRoadmapExitVersionedSizeBoundedV2 IdentityImplementationRoadmapExitIDV2 = "resolver_updates_are_versioned_and_size_bounded"
)

type IdentityRoadmapTaskV2 struct {
	ID       IdentityImplementationRoadmapTaskIDV2
	Evidence []string
}

type IdentityRoadmapExitCriterionV2 struct {
	ID       IdentityImplementationRoadmapExitIDV2
	Evidence []string
}

type IdentityRoadmapPhaseV2 struct {
	ID           IdentityImplementationRoadmapPhaseIDV2
	Title        string
	Tasks        []IdentityRoadmapTaskV2
	ExitCriteria []IdentityRoadmapExitCriterionV2
}

type IdentityImplementationRoadmapV2 struct {
	Phases      []IdentityRoadmapPhaseV2
	RoadmapHash string
}

func DefaultIdentityImplementationRoadmapV2() IdentityImplementationRoadmapV2 {
	roadmap := IdentityImplementationRoadmapV2{Phases: []IdentityRoadmapPhaseV2{
		{
			ID:    IdentityRoadmapPhase0SpecVectorsV2,
			Title: "Specification and Test Vectors",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskCanonicalNameNormalizationV2, Evidence: []string{"x/identity/types/validation_v2.go:NormalizeAETDomainVersioned", "x/identity/types/validation_v2_test.go:TestNameNormalizationV2ValidAndInvalidVectors"}},
				{ID: IdentityRoadmapTaskDomainProofHashFormatsV2, Evidence: []string{"x/identity/types/domain_v2.go:DomainRecordV2NameHash", "x/identity/types/proof_format_v2.go:ComputeIdentityResolutionProofCommitmentHashV2", "x/identity/types/proof_format_v2.go:ComputeRecursiveResolutionProofCommitmentHashV2"}},
				{ID: IdentityRoadmapTaskProtobufStateSchemasV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultResolverModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultSubdomainModuleBreakdownV2"}},
				{ID: IdentityRoadmapTaskStoreV2KeyLayoutV2, Evidence: []string{"x/identity/types/storev2.go:IdentityStoreV2SpecDomainKey", "x/identity/types/storev2.go:IdentityStoreV2SpecResolutionProofReadAccessSet"}},
				{ID: IdentityRoadmapTaskGovernanceParamsV2, Evidence: []string{"x/identity/types/governance_params_v2.go:DefaultIdentityGovernanceParamsV2", "x/identity/types/governance_params_v2.go:ValidateIdentityGovernanceParamsV2"}},
				{ID: IdentityRoadmapTaskResolutionProofVectorsV2, Evidence: []string{"x/identity/types/proof_format_v2.go:BuildIdentityResolutionProofFormatV2", "x/identity/types/proof_format_v2.go:ValidateIdentityResolutionProofFormatV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment"}},
				{ID: IdentityRoadmapTaskLifecycleVectorsV2, Evidence: []string{"x/identity/types/lifecycle_state_machine_v2.go:ApplyDomainLifecycleTransitionV2", "x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2RegistrationRenewalGraceAndRelease"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitSignableHashableVectorsV2, Evidence: []string{"x/identity/types/proof_format_v2.go:ComputeIdentityResolutionProofCommitmentHashV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment"}},
				{ID: IdentityRoadmapExitLifecycleDeterminismV2, Evidence: []string{"x/identity/types/lifecycle_state_machine_v2.go:ApplyDomainLifecycleTransitionV2", "x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2AuctionAlternative"}},
				{ID: IdentityRoadmapExitStorePrefixesFinalizedV2, Evidence: []string{"x/identity/types/storev2.go:IdentityStoreV2SpecDomainKey", "x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPrimaryKeyLayout"}},
			},
		},
		{
			ID:    IdentityRoadmapPhase1CoreActivationV2,
			Title: "Core Registry Activation",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskIdentityCoreModuleV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/identity_core_module_breakdown_v2_test.go:TestIdentityCoreModuleBreakdownV2CoversSection131"}},
				{ID: IdentityRoadmapTaskCoreLifecycleV2, Evidence: []string{"x/identity/types/anti_squatting_v2.go:ReleaseExpiredIdentityDomainV2", "x/identity/types/spec_state.go:CommitDomainRegistration", "x/identity/types/spec_state.go:RevealRegisterDomain", "x/identity/types/spec_state.go:TransferDomainNFT"}},
				{ID: IdentityRoadmapTaskNFTBindingV2, Evidence: []string{"x/identity/types/nft_binding.go:TransferDomainNFTBindingAtomic", "x/identity/types/validation_v2.go:TransferDomainNFTBindingWithInvariantsV2"}},
				{ID: IdentityRoadmapTaskOwnerExpiryIndexesV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:BuildIdentityCoreDerivedIndexesV2", "x/identity/types/identity_core_module_breakdown_v2.go:ValidateIdentityCoreStoreV2IndexesV2"}},
				{ID: IdentityRoadmapTaskCoreQueriesV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/query_v2.go:NewIdentityQueryServiceV2"}},
				{ID: IdentityRoadmapTaskInvariantChecksV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:ValidateIdentityCoreModuleInvariantsV2", "x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitOnChainOwnershipV2, Evidence: []string{"x/identity/types/spec_state.go:RevealRegisterDomain", "x/identity/types/spec_test.go:TestIdentitySpecRegisterAETDomain"}},
				{ID: IdentityRoadmapExitAtomicNFTOwnershipV2, Evidence: []string{"x/identity/types/nft_binding.go:TransferDomainNFTBindingAtomic", "x/identity/types/ownership_consistency_v2_test.go:TestIdentityNFTTransferHooksUpdateOrRejectAtomicallyV2"}},
				{ID: IdentityRoadmapExitExportImportRegistryV2, Evidence: []string{"x/identity/types/spec_state.go:ImportIdentityState", "x/identity/types/spec_test.go:TestIdentitySpecExportImportPreservesDomainLifecycle"}},
			},
		},
		{
			ID:    IdentityRoadmapPhase2UnifiedResolverV2,
			Title: "Unified Resolver",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskResolverModuleV2, Evidence: []string{"x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultResolverModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2_test.go:TestResolverModuleBreakdownV2CoversSection132"}},
				{ID: IdentityRoadmapTaskPrimaryResolutionV2, Evidence: []string{"x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2", "x/identity/types/spec_state.go:ResolveIdentityAddress"}},
				{ID: IdentityRoadmapTaskContractTargetsV2, Evidence: []string{"x/identity/types/resolution_v2.go:NewContractTargetV2", "x/identity/types/routing_integration_module_breakdown_v2.go:BuildRoutingIntegrationContractInvocationMappingV2"}},
				{ID: IdentityRoadmapTaskServiceEndpointsV2, Evidence: []string{"x/identity/types/service_interface_mapping_v2.go:BuildIdentityServiceDiscoveryV2", "x/identity/types/service_interface_mapping_v2.go:DefaultIdentityServiceEndpointTypeRegistryV2"}},
				{ID: IdentityRoadmapTaskInterfaceDescriptorsV2, Evidence: []string{"x/identity/types/resolution_v2.go:InterfaceDescriptorHashV2", "x/identity/types/service_interface_mapping_v2.go:BuildIdentityInterfaceSchemaMappingV2"}},
				{ID: IdentityRoadmapTaskRoutingMetadataV2, Evidence: []string{"x/identity/types/routing_integration_module_breakdown_v2.go:BuildRoutingIntegrationWalletSDKHelperV2", "x/identity/types/routing_integration_module_breakdown_v2.go:DefaultRoutingIntegrationModuleBreakdownV2"}},
				{ID: IdentityRoadmapTaskReverseResolutionV2, Evidence: []string{"x/identity/types/resolution_v2.go:ValidateReverseResolutionRecordV2", "x/identity/types/resolution_v2.go:VerifyReverseResolutionTransactionV2"}},
				{ID: IdentityRoadmapTaskBatchResolverUpdatesV2, Evidence: []string{"x/identity/types/batch_resolver_v2.go:ExecuteBatchResolverUpdatesV2", "x/identity/types/batch_resolver_v2.go:ValidateBatchResolverUpdateResponseV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitUnifiedTargetsV2, Evidence: []string{"x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2", "x/identity/types/v2_test.go:TestUnifiedResolverMetadataAndNamedExecution"}},
				{ID: IdentityRoadmapExitReverseConsistencyV2, Evidence: []string{"x/identity/types/resolution_v2.go:ValidateReverseResolutionRecordV2", "x/identity/types/resolution_v2_test.go:TestReverseResolutionVerificationTransactionV2ChecksVersionAndForward"}},
				{ID: IdentityRoadmapExitVersionedSizeBoundedV2, Evidence: []string{"x/identity/types/batch_resolver_v2.go:ExecuteBatchResolverUpdatesV2", "x/identity/types/resolution_v2.go:ValidateUnifiedResolutionRecordV2"}},
			},
		},
	}}
	roadmap.RoadmapHash = ComputeIdentityImplementationRoadmapHashV2(roadmap)
	return roadmap
}

func ValidateIdentityImplementationRoadmapV2(roadmap IdentityImplementationRoadmapV2) error {
	required := requiredIdentityRoadmapPhaseIDsV2()
	if len(roadmap.Phases) != len(required) {
		return fmt.Errorf("identity implementation roadmap must define %d phases", len(required))
	}
	for i, phaseID := range required {
		if roadmap.Phases[i].ID != phaseID {
			return fmt.Errorf("identity implementation roadmap phase %d must be %s", i, phaseID)
		}
		if err := validateIdentityRoadmapPhaseV2(roadmap.Phases[i]); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash == "" || roadmap.RoadmapHash != ComputeIdentityImplementationRoadmapHashV2(roadmap) {
		return errors.New("identity implementation roadmap hash mismatch")
	}
	return nil
}

func ComputeIdentityImplementationRoadmapHashV2(roadmap IdentityImplementationRoadmapV2) string {
	parts := []string{"identity-implementation-roadmap-v2"}
	for _, phase := range roadmap.Phases {
		parts = append(parts, string(phase.ID), phase.Title)
		for _, task := range phase.Tasks {
			parts = append(parts, string(task.ID))
			parts = append(parts, sortedBreakdownStringsV2(task.Evidence)...)
		}
		for _, criterion := range phase.ExitCriteria {
			parts = append(parts, string(criterion.ID))
			parts = append(parts, sortedBreakdownStringsV2(criterion.Evidence)...)
		}
	}
	return identityHash(parts...)
}

func requiredIdentityRoadmapPhaseIDsV2() []IdentityImplementationRoadmapPhaseIDV2 {
	return []IdentityImplementationRoadmapPhaseIDV2{
		IdentityRoadmapPhase0SpecVectorsV2,
		IdentityRoadmapPhase1CoreActivationV2,
		IdentityRoadmapPhase2UnifiedResolverV2,
	}
}

func validateIdentityRoadmapPhaseV2(phase IdentityRoadmapPhaseV2) error {
	if phase.Title == "" {
		return fmt.Errorf("identity implementation roadmap phase %s title is required", phase.ID)
	}
	if !identityRoadmapTasksEqualV2(phase.Tasks, requiredIdentityRoadmapTasksV2(phase.ID)) {
		return fmt.Errorf("identity implementation roadmap phase %s tasks mismatch", phase.ID)
	}
	if !identityRoadmapExitsEqualV2(phase.ExitCriteria, requiredIdentityRoadmapExitsV2(phase.ID)) {
		return fmt.Errorf("identity implementation roadmap phase %s exit criteria mismatch", phase.ID)
	}
	for _, task := range phase.Tasks {
		if err := validateIdentityRoadmapEvidenceV2("task", string(task.ID), task.Evidence); err != nil {
			return err
		}
	}
	for _, criterion := range phase.ExitCriteria {
		if err := validateIdentityRoadmapEvidenceV2("exit criterion", string(criterion.ID), criterion.Evidence); err != nil {
			return err
		}
	}
	return nil
}

func requiredIdentityRoadmapTasksV2(phase IdentityImplementationRoadmapPhaseIDV2) []IdentityImplementationRoadmapTaskIDV2 {
	switch phase {
	case IdentityRoadmapPhase0SpecVectorsV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskCanonicalNameNormalizationV2, IdentityRoadmapTaskDomainProofHashFormatsV2, IdentityRoadmapTaskProtobufStateSchemasV2, IdentityRoadmapTaskStoreV2KeyLayoutV2, IdentityRoadmapTaskGovernanceParamsV2, IdentityRoadmapTaskResolutionProofVectorsV2, IdentityRoadmapTaskLifecycleVectorsV2}
	case IdentityRoadmapPhase1CoreActivationV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskIdentityCoreModuleV2, IdentityRoadmapTaskCoreLifecycleV2, IdentityRoadmapTaskNFTBindingV2, IdentityRoadmapTaskOwnerExpiryIndexesV2, IdentityRoadmapTaskCoreQueriesV2, IdentityRoadmapTaskInvariantChecksV2}
	case IdentityRoadmapPhase2UnifiedResolverV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskResolverModuleV2, IdentityRoadmapTaskPrimaryResolutionV2, IdentityRoadmapTaskContractTargetsV2, IdentityRoadmapTaskServiceEndpointsV2, IdentityRoadmapTaskInterfaceDescriptorsV2, IdentityRoadmapTaskRoutingMetadataV2, IdentityRoadmapTaskReverseResolutionV2, IdentityRoadmapTaskBatchResolverUpdatesV2}
	default:
		return nil
	}
}

func requiredIdentityRoadmapExitsV2(phase IdentityImplementationRoadmapPhaseIDV2) []IdentityImplementationRoadmapExitIDV2 {
	switch phase {
	case IdentityRoadmapPhase0SpecVectorsV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitSignableHashableVectorsV2, IdentityRoadmapExitLifecycleDeterminismV2, IdentityRoadmapExitStorePrefixesFinalizedV2}
	case IdentityRoadmapPhase1CoreActivationV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitOnChainOwnershipV2, IdentityRoadmapExitAtomicNFTOwnershipV2, IdentityRoadmapExitExportImportRegistryV2}
	case IdentityRoadmapPhase2UnifiedResolverV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitUnifiedTargetsV2, IdentityRoadmapExitReverseConsistencyV2, IdentityRoadmapExitVersionedSizeBoundedV2}
	default:
		return nil
	}
}

func validateIdentityRoadmapEvidenceV2(kind string, id string, evidence []string) error {
	if len(evidence) == 0 {
		return fmt.Errorf("identity implementation roadmap %s %s evidence is required", kind, id)
	}
	sorted := append([]string(nil), evidence...)
	sort.Strings(sorted)
	for i, ref := range sorted {
		if ref == "" || !identityRoadmapEvidenceReferenceHasFunctionV2(ref) {
			return fmt.Errorf("identity implementation roadmap %s %s invalid evidence reference %q", kind, id, ref)
		}
		if ref != evidence[i] {
			return fmt.Errorf("identity implementation roadmap %s %s evidence must be sorted", kind, id)
		}
		if i > 0 && sorted[i-1] == ref {
			return fmt.Errorf("duplicate identity implementation roadmap evidence reference %q", ref)
		}
	}
	return nil
}

func identityRoadmapEvidenceReferenceHasFunctionV2(ref string) bool {
	return strings.Contains(ref, ".go:")
}

func identityRoadmapTasksEqualV2(tasks []IdentityRoadmapTaskV2, required []IdentityImplementationRoadmapTaskIDV2) bool {
	if len(tasks) != len(required) {
		return false
	}
	for i := range required {
		if tasks[i].ID != required[i] {
			return false
		}
	}
	return true
}

func identityRoadmapExitsEqualV2(exits []IdentityRoadmapExitCriterionV2, required []IdentityImplementationRoadmapExitIDV2) bool {
	if len(exits) != len(required) {
		return false
	}
	for i := range required {
		if exits[i].ID != required[i] {
			return false
		}
	}
	return true
}
