package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMImplementationRoadmapMatchesSection19(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	require.NoError(t, roadmap.Validate())
	require.Equal(t, "AVM implementation roadmap", roadmap.RoadmapName)
	require.Equal(t, ComputeAVMImplementationRoadmapHash(roadmap), roadmap.RoadmapHash)
	require.Len(t, roadmap.Phases, 5)

	byPhase := map[AVMImplementationRoadmapPhaseID]AVMImplementationRoadmapPhase{}
	for _, phase := range roadmap.Phases {
		require.NoError(t, phase.Validate())
		require.True(t, phase.ConsensusCritical)
		byPhase[phase.PhaseID] = phase
	}
	require.Equal(t, "Specification and Test Vectors", byPhase[AVMRoadmapPhase0].Name)
	require.Equal(t, "Sync Router", byPhase[AVMRoadmapPhase1].Name)
	require.Equal(t, "Async Engine", byPhase[AVMRoadmapPhase2].Name)
	require.Equal(t, "Cross-Zone Routing", byPhase[AVMRoadmapPhase3].Name)
	require.Equal(t, "Actor Runtime", byPhase[AVMRoadmapPhase4].Name)
}

func TestAVMRoadmapPhase0DefinesSpecAndTestVectors(t *testing.T) {
	phase, err := NewAVMImplementationRoadmapPhase(AVMImplementationRoadmapPhase{
		PhaseID: AVMRoadmapPhase0,
		Name:    "Specification and Test Vectors",
		Tasks: []AVMImplementationRoadmapTask{
			AVMRoadmapTaskCanonicalAsyncMessageEncoding,
			AVMRoadmapTaskMessageIDDerivation,
			AVMRoadmapTaskDeterministicQueueSortKey,
			AVMRoadmapTaskExecutionReceiptSchema,
			AVMRoadmapTaskAVMRootSchema,
			AVMRoadmapTaskGasPolicySchema,
			AVMRoadmapTaskInterfaceDescriptorSchema,
		},
		ExitCriteria: []AVMImplementationExitCriterion{
			AVMRoadmapExitSignableHashableObjectsHaveTestVectors,
			AVMRoadmapExitQueueOrderingTestCovered,
			AVMRoadmapExitRootEncodingFixed,
		},
		TestVectorTargets: []AVMTestVectorTarget{
			AVMRoadmapVectorAsyncMessageEncoding,
			AVMRoadmapVectorMessageIDDerivation,
			AVMRoadmapVectorDeterministicQueueSort,
			AVMRoadmapVectorExecutionReceiptSchema,
			AVMRoadmapVectorAVMRootSchema,
			AVMRoadmapVectorGasPolicySchema,
			AVMRoadmapVectorInterfaceDescriptorSchema,
		},
		ConsensusCritical: true,
	})
	require.NoError(t, err)
	require.NoError(t, phase.Validate())

	require.ElementsMatch(t, []AVMImplementationRoadmapTask{
		AVMRoadmapTaskCanonicalAsyncMessageEncoding,
		AVMRoadmapTaskMessageIDDerivation,
		AVMRoadmapTaskDeterministicQueueSortKey,
		AVMRoadmapTaskExecutionReceiptSchema,
		AVMRoadmapTaskAVMRootSchema,
		AVMRoadmapTaskGasPolicySchema,
		AVMRoadmapTaskInterfaceDescriptorSchema,
	}, phase.Tasks)
	require.ElementsMatch(t, []AVMImplementationExitCriterion{
		AVMRoadmapExitSignableHashableObjectsHaveTestVectors,
		AVMRoadmapExitQueueOrderingTestCovered,
		AVMRoadmapExitRootEncodingFixed,
	}, phase.ExitCriteria)
	require.ElementsMatch(t, []AVMTestVectorTarget{
		AVMRoadmapVectorAsyncMessageEncoding,
		AVMRoadmapVectorMessageIDDerivation,
		AVMRoadmapVectorDeterministicQueueSort,
		AVMRoadmapVectorExecutionReceiptSchema,
		AVMRoadmapVectorAVMRootSchema,
		AVMRoadmapVectorGasPolicySchema,
		AVMRoadmapVectorInterfaceDescriptorSchema,
	}, phase.TestVectorTargets)
}

func TestAVMRoadmapPhase1DefinesSyncRouterMilestone(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	phase := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase1)

	require.ElementsMatch(t, []AVMImplementationRoadmapTask{
		AVMRoadmapTaskRouterSkeleton,
		AVMRoadmapTaskSyncEngineWrapper,
		AVMRoadmapTaskSyncReceipts,
		AVMRoadmapTaskZoneRouteDescriptors,
		AVMRoadmapTaskAVMRootPlaceholder,
	}, phase.Tasks)
	require.ElementsMatch(t, []AVMImplementationExitCriterion{
		AVMRoadmapExitExistingCallsRepresentedAsRoutedSync,
		AVMRoadmapExitDeterministicSyncReceipts,
	}, phase.ExitCriteria)
	require.Empty(t, phase.TestVectorTargets)
}

func TestAVMRoadmapPhase2DefinesAsyncEngineMilestone(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	phase := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase2)

	require.ElementsMatch(t, []AVMImplementationRoadmapTask{
		AVMRoadmapTaskAsyncMessageStore,
		AVMRoadmapTaskZoneQueues,
		AVMRoadmapTaskDelayedQueue,
		AVMRoadmapTaskRetryQueue,
		AVMRoadmapTaskDeadLetterQueue,
		AVMRoadmapTaskReplayTombstones,
		AVMRoadmapTaskQueueRoots,
	}, phase.Tasks)
	require.ElementsMatch(t, []AVMImplementationExitCriterion{
		AVMRoadmapExitAsyncScheduledLaterBlocks,
		AVMRoadmapExitExpiredMessagesDoNotRun,
		AVMRoadmapExitRetryDeadLetterDeterminism,
	}, phase.ExitCriteria)
	require.Empty(t, phase.TestVectorTargets)
}

func TestAVMRoadmapPhase3DefinesCrossZoneRoutingMilestone(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	phase := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase3)

	require.ElementsMatch(t, []AVMImplementationRoadmapTask{
		AVMRoadmapTaskZoneMetadata,
		AVMRoadmapTaskZoneMessageFilters,
		AVMRoadmapTaskCrossZoneInboxOutboxRoot,
		AVMRoadmapTaskBounceSystem,
		AVMRoadmapTaskCrossZoneValueAccounting,
		AVMRoadmapTaskCrossZoneProofQueries,
	}, phase.Tasks)
	require.ElementsMatch(t, []AVMImplementationExitCriterion{
		AVMRoadmapExitZoneAToZoneBAsync,
		AVMRoadmapExitFailedCrossZoneTerminal,
		AVMRoadmapExitCrossZoneReceiptsProofable,
	}, phase.ExitCriteria)
	require.Empty(t, phase.TestVectorTargets)
}

func TestAVMRoadmapPhase4DefinesActorRuntimeMilestone(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	phase := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase4)

	require.ElementsMatch(t, []AVMImplementationRoadmapTask{
		AVMRoadmapTaskActorRecords,
		AVMRoadmapTaskActorMailboxes,
		AVMRoadmapTaskActorHandlerDispatch,
		AVMRoadmapTaskActorStateIsolation,
		AVMRoadmapTaskActorReceiptEmission,
		AVMRoadmapTaskActorStateProofQuery,
	}, phase.Tasks)
	require.ElementsMatch(t, []AVMImplementationExitCriterion{
		AVMRoadmapExitActorMailboxSerialExecution,
		AVMRoadmapExitActorStateIsolation,
		AVMRoadmapExitActorDeterministicReceipts,
	}, phase.ExitCriteria)
	require.Empty(t, phase.TestVectorTargets)
}

func TestAVMRoadmapRejectsMissingCrossOwnedAndHashMismatch(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)

	missingPhase := roadmap
	missingPhase.Phases = missingPhase.Phases[:len(missingPhase.Phases)-1]
	missingPhase.RoadmapHash = ComputeAVMImplementationRoadmapHash(missingPhase)
	require.ErrorContains(t, missingPhase.Validate(), "every section 19 phase")

	mutated := roadmap
	mutated.Phases[0].Name = "Changed"
	require.ErrorContains(t, mutated.Validate(), "phase hash mismatch")

	phase0 := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase0)
	phase0.Tasks = removeAVMRoadmapTaskForTest(phase0.Tasks, AVMRoadmapTaskMessageIDDerivation)
	phase0.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase0)
	require.ErrorContains(t, phase0.Validate(), "task")

	phase1 := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase1)
	phase1.Tasks = append(phase1.Tasks, AVMRoadmapTaskAsyncMessageStore)
	phase1.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase1)
	require.ErrorContains(t, phase1.Validate(), "task")

	phase0 = roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase0)
	phase0.TestVectorTargets = removeAVMTestVectorTargetForTest(phase0.TestVectorTargets, AVMRoadmapVectorAVMRootSchema)
	phase0.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase0)
	require.ErrorContains(t, phase0.Validate(), "test vector target")

	phase2 := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase2)
	phase2.TestVectorTargets = append(phase2.TestVectorTargets, AVMRoadmapVectorAsyncMessageEncoding)
	phase2.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase2)
	require.ErrorContains(t, phase2.Validate(), "test vector target")

	phase3 := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase3)
	phase3.Tasks = append(phase3.Tasks, AVMRoadmapTaskActorRecords)
	phase3.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase3)
	require.ErrorContains(t, phase3.Validate(), "task")

	phase4 := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase4)
	phase4.ExitCriteria = removeAVMRoadmapExitCriterionForTest(phase4.ExitCriteria, AVMRoadmapExitActorStateIsolation)
	phase4.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase4)
	require.ErrorContains(t, phase4.Validate(), "exit criterion")
}

func TestAVMRoadmapRejectsNonConsensusCriticalPhase(t *testing.T) {
	roadmap, err := DefaultAVMImplementationRoadmap()
	require.NoError(t, err)
	phase := roadmapPhaseForTest(t, roadmap, AVMRoadmapPhase1)
	phase.ConsensusCritical = false
	phase.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase)
	require.ErrorContains(t, phase.Validate(), "consensus-critical")
}

func roadmapPhaseForTest(t *testing.T, roadmap AVMImplementationRoadmap, phaseID AVMImplementationRoadmapPhaseID) AVMImplementationRoadmapPhase {
	t.Helper()
	for _, phase := range roadmap.Phases {
		if phase.PhaseID == phaseID {
			return phase
		}
	}
	t.Fatalf("missing roadmap phase %s", phaseID)
	return AVMImplementationRoadmapPhase{}
}

func removeAVMRoadmapTaskForTest(tasks []AVMImplementationRoadmapTask, target AVMImplementationRoadmapTask) []AVMImplementationRoadmapTask {
	out := make([]AVMImplementationRoadmapTask, 0, len(tasks))
	for _, task := range tasks {
		if task != target {
			out = append(out, task)
		}
	}
	return out
}

func removeAVMRoadmapExitCriterionForTest(criteria []AVMImplementationExitCriterion, target AVMImplementationExitCriterion) []AVMImplementationExitCriterion {
	out := make([]AVMImplementationExitCriterion, 0, len(criteria))
	for _, criterion := range criteria {
		if criterion != target {
			out = append(out, criterion)
		}
	}
	return out
}

func removeAVMTestVectorTargetForTest(targets []AVMTestVectorTarget, target AVMTestVectorTarget) []AVMTestVectorTarget {
	out := make([]AVMTestVectorTarget, 0, len(targets))
	for _, vector := range targets {
		if vector != target {
			out = append(out, vector)
		}
	}
	return out
}
