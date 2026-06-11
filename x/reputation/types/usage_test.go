package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestLimitsForReputation(t *testing.T) {
	policy := DefaultUsagePolicy()

	low := LimitsForReputation(10, policy)
	require.Equal(t, uint32(1), low.TxRateLimit)
	require.Equal(t, uint32(1), low.AsyncQueueQuota)
	require.Equal(t, uint64(4), low.MemoByteCost)
	require.Equal(t, uint64(40), low.StorageByteCost)
	require.Equal(t, uint32(0), low.ContractDeploysPerEpoch)

	high := LimitsForReputation(95, policy)
	require.Equal(t, uint32(250), high.TxRateLimit)
	require.Equal(t, uint32(128), high.AsyncQueueQuota)
	require.Equal(t, uint64(1), high.MemoByteCost)
	require.Equal(t, uint64(7), high.StorageByteCost)
	require.Equal(t, uint32(10), high.ContractDeploysPerEpoch)
}

func TestValidateTxUsageRateLimitAndFeeRequired(t *testing.T) {
	policy := DefaultUsagePolicy()
	record := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 10})

	require.NoError(t, ValidateTxUsage(record, 0, true, true, policy))
	require.ErrorContains(t, ValidateTxUsage(record, 1, true, true, policy), "tx rate limit")
	require.ErrorContains(t, ValidateTxUsage(record, 0, false, true, policy), "required protocol fee")
	require.ErrorContains(t, ValidateTxUsage(record, 0, true, false, policy), "base transaction validation")

	elite := ApplyComputedScore(ReputationRecord{Account: addr(2), AgeScore: 100})
	require.ErrorContains(t, ValidateTxUsage(elite, 0, false, true, policy), "required protocol fee")
}

func TestValidateAsyncQueueUsage(t *testing.T) {
	policy := DefaultUsagePolicy()
	record := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 10})

	require.NoError(t, ValidateAsyncQueueUsage(record, 0, true, policy))
	require.ErrorContains(t, ValidateAsyncQueueUsage(record, 1, true, policy), "async queue quota")
	require.ErrorContains(t, ValidateAsyncQueueUsage(record, 0, false, policy), "required protocol fee")
}

func TestValidateAccessOperationThresholdOrDeposit(t *testing.T) {
	policy := DefaultUsagePolicy()
	low := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 20})
	normal := ApplyComputedScore(ReputationRecord{Account: addr(2), AgeScore: 50})

	require.ErrorContains(t, ValidateAccessOperation(OperationContractDeployment, low, sdkmath.ZeroInt(), policy), "requires reputation")
	require.NoError(t, ValidateAccessOperation(OperationContractDeployment, low, policy.ContractDeployDeposit, policy))
	require.NoError(t, ValidateAccessOperation(OperationContractDeployment, normal, sdkmath.ZeroInt(), policy))

	require.ErrorContains(t, ValidateAccessOperation("unknown", normal, sdkmath.ZeroInt(), policy), "unknown")
	require.False(t, IsDirectReputationPurchaseAllowed())
}

func TestPriorityForReputationDeterministicAndBounded(t *testing.T) {
	require.Equal(t, PriorityKey{Weight: RestrictedPriorityWeight, TxIndex: 7}, PriorityForReputation(0, 7))
	require.Equal(t, PriorityKey{Weight: ElitePriorityWeight, TxIndex: 7}, PriorityForReputation(100, 7))
	require.Less(t, PriorityWeight(50), PriorityWeight(95))
	require.LessOrEqual(t, PriorityWeight(100), ElitePriorityWeight)
}

func TestContractExecutionOutcome(t *testing.T) {
	record := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 50})

	updated := ApplyContractExecutionOutcome(record, true, 10)
	require.Equal(t, uint16(1), updated.ContractScore)
	require.Equal(t, uint64(10), updated.LastUpdatedEpoch)
	require.Equal(t, ComputeScore(updated), updated.Score)

	failed := ApplyContractExecutionOutcome(updated, false, 11)
	require.Equal(t, uint16(5), failed.FailedTxPenalty)
	require.Equal(t, uint64(11), failed.LastUpdatedEpoch)
	require.Equal(t, ComputeScore(failed), failed.Score)
}
