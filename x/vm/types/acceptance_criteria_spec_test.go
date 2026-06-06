package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMAcceptanceCriteriaSpecMatchesSection22(t *testing.T) {
	spec, err := DefaultAVMAcceptanceCriteriaSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Equal(t, ComputeAVMAcceptanceCriteriaSpecHash(spec), spec.SpecHash)
	require.Len(t, spec.Criteria, 13)
	require.ElementsMatch(t, []AVMAcceptanceCriterion{
		AVMAcceptanceSyncExecutionWrapsMsgServer,
		AVMAcceptanceAsyncLifecycleComplete,
		AVMAcceptanceQueueOrderingDeterministicTested,
		AVMAcceptanceCrossZoneQueueReceiptRoots,
		AVMAcceptanceActorMailboxIsolation,
		AVMAcceptanceContinuationsPauseResume,
		AVMAcceptanceContractBackendsExplicitInterface,
		AVMAcceptanceGasModelComplete,
		AVMAcceptanceInterfaceRegistryComplete,
		AVMAcceptanceRootCommitmentsComplete,
		AVMAcceptanceReplayProtection,
		AVMAcceptanceStoreV2PrefixProofQueryable,
		AVMAcceptanceBlockSTMConflictStrategyBenchmarked,
	}, spec.Criteria)
}

func TestAVMAcceptanceCriteriaRejectsMissingDuplicateUnknownAndHashMismatch(t *testing.T) {
	spec, err := DefaultAVMAcceptanceCriteriaSpec()
	require.NoError(t, err)

	missing := spec
	missing.Criteria = missing.Criteria[:1]
	missing.SpecHash = ComputeAVMAcceptanceCriteriaSpecHash(missing)
	require.ErrorContains(t, missing.Validate(), "every section 22 criterion")

	spec, err = DefaultAVMAcceptanceCriteriaSpec()
	require.NoError(t, err)
	duplicate := spec
	duplicate.Criteria[1] = duplicate.Criteria[0]
	duplicate.SpecHash = ComputeAVMAcceptanceCriteriaSpecHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	spec, err = DefaultAVMAcceptanceCriteriaSpec()
	require.NoError(t, err)
	unknown := spec
	unknown.Criteria[0] = AVMAcceptanceCriterion("unknown_criterion")
	unknown.SpecHash = ComputeAVMAcceptanceCriteriaSpecHash(unknown)
	require.ErrorContains(t, unknown.Validate(), "invalid")

	spec, err = DefaultAVMAcceptanceCriteriaSpec()
	require.NoError(t, err)
	hashMismatch := spec
	hashMismatch.SpecHash = "0000000000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, hashMismatch.Validate(), "hash mismatch")
}
