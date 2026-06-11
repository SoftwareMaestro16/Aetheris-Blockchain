package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestIdentityReputationNonGatingOperationsAndMinimumStake(t *testing.T) {
	rep := NewIdentityReputation(addressing.FormatAccAddress(bytes20(0x91)))

	require.NoError(t, ValidateIdentityReputation(rep))
	require.NoError(t, ValidateNonGatingEnforcement())

	for _, op := range []string{
		"contract_deployment",
		"contract_execution",
		"basic_transaction",
		"pool_staking",
	} {
		require.True(t, CanLowReputationPerformOperation(op), op)
		require.False(t, IsOperationGated(op), op)
	}

	for _, op := range []string{
		"tx_priority",
		"async_queue_ordering",
		"resource_scheduling",
		"fee_discount",
		"liquid_staking_yield",
	} {
		require.False(t, CanLowReputationPerformOperation(op), op)
		require.True(t, IsOperationGated(op), op)
	}

	require.True(t, CanStakeMinimum(rep, MinStakeAmountForReputation))
	require.False(t, CanStakeMinimum(rep, MinStakeAmountForReputation-1))
}

func TestIdentityReputationExportImportAndSingleQuery(t *testing.T) {
	rep := NewIdentityReputation(addressing.FormatAccAddress(bytes20(0x92)))
	rep.RecordSuccessfulTx(10)
	rep.RecordFailedTx(11)
	rep.RecordContractInteraction(12)
	rep.RecordContractFailure(13)
	rep.RecordDomainRegistration(14)
	rep.RecordUptime(200, 15)
	rep.RecordStakeTime(365*24*3600, 16)

	rep.Score = ComputeIdentityScore(rep)
	rep.Confidence = ComputeConfidence(rep)

	require.NoError(t, ValidateIdentityReputation(rep))
	require.Greater(t, rep.Confidence, ConfidenceDefault)
	require.LessOrEqual(t, rep.Score, IdentityScoreMax)

	claim := rep.ExportClaim(99)
	imported, err := ImportReputationFromClaim(claim)
	require.NoError(t, err)
	require.Equal(t, rep.Account, imported.Account)
	require.Equal(t, rep.Score, imported.Score)
	require.Equal(t, rep.Confidence, imported.Confidence)
	require.Equal(t, rep.StakeTimeAccumulator, imported.StakeTimeAccumulator)
	require.Equal(t, rep.DecayEpoch, imported.DecayEpoch)

	validatorScore := NewValidatorScore("val-1")
	serviceTrust := NewServiceTrustScore("svc-1")
	response := QueryIdentityReputation(imported, validatorScore, serviceTrust)
	require.Equal(t, imported.Account, response.Reputation.Account)
	require.Equal(t, ClassifyReputationLevel(imported.Score), response.Level)
	require.Equal(t, GetIdentityProgressiveLimits(response.Level), response.ProgressiveLimits)
	require.NotNil(t, response.ValidatorScore)
	require.NotNil(t, response.ServiceTrust)
	require.Equal(t, validatorScore.ValidatorAddress, response.ValidatorScore.ValidatorAddress)
	require.Equal(t, serviceTrust.ServiceAddress, response.ServiceTrust.ServiceAddress)
}

func TestIdentityReputationRejectsNonIdentityContractSubjects(t *testing.T) {
	require.ErrorContains(t, ValidateNoContractReputation("AEcontract123"), "do not have persistent reputation")
	require.NoError(t, ValidateReputationAccountAddress("AEAAAQAAA"))
	require.NoError(t, ValidateReputationAccountAddress("4:0000000000000000000000000000000000000000000000000000000000000001"))
	require.Error(t, ValidateReputationAccountAddress("cosmos1bad"))
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
