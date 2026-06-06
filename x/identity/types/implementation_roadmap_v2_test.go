package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityImplementationRoadmapV2CoversPhases0To2(t *testing.T) {
	roadmap := DefaultIdentityImplementationRoadmapV2()
	require.NoError(t, ValidateIdentityImplementationRoadmapV2(roadmap))
	require.Len(t, roadmap.Phases, 3)
	require.NotEmpty(t, roadmap.RoadmapHash)

	phase0 := roadmap.Phases[0]
	require.Equal(t, IdentityRoadmapPhase0SpecVectorsV2, phase0.ID)
	requireRoadmapTasksV2(t, phase0, []IdentityImplementationRoadmapTaskIDV2{
		IdentityRoadmapTaskCanonicalNameNormalizationV2,
		IdentityRoadmapTaskDomainProofHashFormatsV2,
		IdentityRoadmapTaskProtobufStateSchemasV2,
		IdentityRoadmapTaskStoreV2KeyLayoutV2,
		IdentityRoadmapTaskGovernanceParamsV2,
		IdentityRoadmapTaskResolutionProofVectorsV2,
		IdentityRoadmapTaskLifecycleVectorsV2,
	})
	requireRoadmapExitsV2(t, phase0, []IdentityImplementationRoadmapExitIDV2{
		IdentityRoadmapExitSignableHashableVectorsV2,
		IdentityRoadmapExitLifecycleDeterminismV2,
		IdentityRoadmapExitStorePrefixesFinalizedV2,
	})

	phase1 := roadmap.Phases[1]
	require.Equal(t, IdentityRoadmapPhase1CoreActivationV2, phase1.ID)
	requireRoadmapTasksV2(t, phase1, []IdentityImplementationRoadmapTaskIDV2{
		IdentityRoadmapTaskIdentityCoreModuleV2,
		IdentityRoadmapTaskCoreLifecycleV2,
		IdentityRoadmapTaskNFTBindingV2,
		IdentityRoadmapTaskOwnerExpiryIndexesV2,
		IdentityRoadmapTaskCoreQueriesV2,
		IdentityRoadmapTaskInvariantChecksV2,
	})
	requireRoadmapExitsV2(t, phase1, []IdentityImplementationRoadmapExitIDV2{
		IdentityRoadmapExitOnChainOwnershipV2,
		IdentityRoadmapExitAtomicNFTOwnershipV2,
		IdentityRoadmapExitExportImportRegistryV2,
	})

	phase2 := roadmap.Phases[2]
	require.Equal(t, IdentityRoadmapPhase2UnifiedResolverV2, phase2.ID)
	requireRoadmapTasksV2(t, phase2, []IdentityImplementationRoadmapTaskIDV2{
		IdentityRoadmapTaskResolverModuleV2,
		IdentityRoadmapTaskPrimaryResolutionV2,
		IdentityRoadmapTaskContractTargetsV2,
		IdentityRoadmapTaskServiceEndpointsV2,
		IdentityRoadmapTaskInterfaceDescriptorsV2,
		IdentityRoadmapTaskRoutingMetadataV2,
		IdentityRoadmapTaskReverseResolutionV2,
		IdentityRoadmapTaskBatchResolverUpdatesV2,
	})
	requireRoadmapExitsV2(t, phase2, []IdentityImplementationRoadmapExitIDV2{
		IdentityRoadmapExitUnifiedTargetsV2,
		IdentityRoadmapExitReverseConsistencyV2,
		IdentityRoadmapExitVersionedSizeBoundedV2,
	})

	requireCoverageReferencesExistV2(t, identityRoadmapEvidenceReferencesV2(roadmap))
}

func TestIdentityImplementationRoadmapV2RejectsMissingTaskBadEvidenceAndHashMismatch(t *testing.T) {
	roadmap := DefaultIdentityImplementationRoadmapV2()

	missingTask := roadmap
	missingTask.Phases = cloneIdentityRoadmapPhasesV2(roadmap.Phases)
	missingTask.Phases[0].Tasks = missingTask.Phases[0].Tasks[:len(missingTask.Phases[0].Tasks)-1]
	missingTask.RoadmapHash = ComputeIdentityImplementationRoadmapHashV2(missingTask)
	require.ErrorContains(t, ValidateIdentityImplementationRoadmapV2(missingTask), "tasks mismatch")

	badEvidence := roadmap
	badEvidence.Phases = cloneIdentityRoadmapPhasesV2(roadmap.Phases)
	badEvidence.Phases[0].Tasks[0].Evidence = []string{"not-a-go-reference"}
	badEvidence.RoadmapHash = ComputeIdentityImplementationRoadmapHashV2(badEvidence)
	require.ErrorContains(t, ValidateIdentityImplementationRoadmapV2(badEvidence), "invalid evidence reference")

	tampered := roadmap
	tampered.Phases = cloneIdentityRoadmapPhasesV2(roadmap.Phases)
	tampered.Phases[2].ExitCriteria[0].Evidence[0] = "x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2Tampered"
	require.ErrorContains(t, ValidateIdentityImplementationRoadmapV2(tampered), "hash mismatch")
}

func requireRoadmapTasksV2(t *testing.T, phase IdentityRoadmapPhaseV2, want []IdentityImplementationRoadmapTaskIDV2) {
	t.Helper()
	require.Len(t, phase.Tasks, len(want))
	for i, expected := range want {
		require.Equal(t, expected, phase.Tasks[i].ID)
		require.NotEmpty(t, phase.Tasks[i].Evidence)
	}
}

func requireRoadmapExitsV2(t *testing.T, phase IdentityRoadmapPhaseV2, want []IdentityImplementationRoadmapExitIDV2) {
	t.Helper()
	require.Len(t, phase.ExitCriteria, len(want))
	for i, expected := range want {
		require.Equal(t, expected, phase.ExitCriteria[i].ID)
		require.NotEmpty(t, phase.ExitCriteria[i].Evidence)
	}
}

func identityRoadmapEvidenceReferencesV2(roadmap IdentityImplementationRoadmapV2) []string {
	refs := make([]string, 0)
	for _, phase := range roadmap.Phases {
		for _, task := range phase.Tasks {
			refs = append(refs, task.Evidence...)
		}
		for _, criterion := range phase.ExitCriteria {
			refs = append(refs, criterion.Evidence...)
		}
	}
	return refs
}

func cloneIdentityRoadmapPhasesV2(in []IdentityRoadmapPhaseV2) []IdentityRoadmapPhaseV2 {
	out := append([]IdentityRoadmapPhaseV2(nil), in...)
	for i := range out {
		out[i].Tasks = append([]IdentityRoadmapTaskV2(nil), out[i].Tasks...)
		for j := range out[i].Tasks {
			out[i].Tasks[j].Evidence = append([]string(nil), out[i].Tasks[j].Evidence...)
		}
		out[i].ExitCriteria = append([]IdentityRoadmapExitCriterionV2(nil), out[i].ExitCriteria...)
		for j := range out[i].ExitCriteria {
			out[i].ExitCriteria[j].Evidence = append([]string(nil), out[i].ExitCriteria[j].Evidence...)
		}
	}
	return out
}
