package params

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultSlashingAccountabilityPolicyMatchesAetraModel(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()

	require.NoError(t, policy.Validate())
	require.Equal(t, SlashingEvidenceStandardCosmos, policy.EvidenceStandard)
	require.True(t, policy.ObjectiveCryptographicEvidenceOnly)
	require.False(t, policy.SubjectiveSlashingAllowed)
	require.Equal(t, int64(500), policy.DoubleSignSlashBps)
	require.GreaterOrEqual(t, policy.DoubleSignSlashBps, DoubleSignSlashMinBps)
	require.LessOrEqual(t, policy.DoubleSignSlashBps, DoubleSignSlashMaxBps)
	require.True(t, policy.DoubleSignJailImmediate)
	require.True(t, policy.DoubleSignPermanentTombstone)
	require.True(t, policy.ConsensusKeyReuseForbidden)
	require.True(t, policy.UsesCosmosSlashingAndEvidence)
	require.True(t, policy.ProgressiveDowntimeEnabled)
	require.True(t, policy.StandardDowntimeStatePreserved)
	require.True(t, policy.CustomDowntimeOverlayRequired)
	require.Equal(t, int64(5), policy.DowntimeFirstSlashBps)
	require.Equal(t, int64(60), policy.DowntimeFirstJailMinutes)
	require.Equal(t, int64(25), policy.DowntimeRepeatSlashBps)
	require.Equal(t, int64(24*60), policy.DowntimeRepeatJailMinutes)
	require.Equal(t, int64(100), policy.DowntimeChronicSlashBps)
	require.Greater(t, policy.DowntimeChronicJailMinutes, policy.DowntimeRepeatJailMinutes)
	require.True(t, policy.DowntimeGovernanceReputationFlag)
	require.True(t, policy.InvalidProposalDeterministicReject)
	require.False(t, policy.InvalidProposalAutoSlash)
	require.True(t, policy.InvalidProposalRepeatEvidenceOnly)
	require.False(t, policy.ProcessProposalExternalInputs)
	require.True(t, policy.ProcessProposalTestsRequired)
	require.Equal(t, int64(25), policy.RepeatedInvalidProposalSlashBps)
	require.Equal(t, int64(24*60), policy.RepeatedInvalidProposalJailMinutes)
	require.True(t, policy.TimestampRejectOutsideBounds)
	require.True(t, policy.TimestampCometBFTCompatible)
	require.False(t, policy.TimestampCustomWallClockLogic)
	require.True(t, policy.TimestampSlashObjectiveEvidenceOnly)
	require.Equal(t, int64(25), policy.TimestampRepeatedViolationsSlashBps)
	require.Equal(t, int64(24*60), policy.TimestampRepeatedViolationsJailMinutes)
	require.Equal(t, int64(120), policy.TimestampMaxForwardDriftSeconds)
}

func TestSlashingAccountabilityPolicyRejectsSubjectiveOrWeakDoubleSignRules(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.DoubleSignSlashBps = 750
	require.NoError(t, policy.Validate())

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ObjectiveCryptographicEvidenceOnly = false
	require.ErrorContains(t, policy.Validate(), "objective cryptographic evidence")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.SubjectiveSlashingAllowed = true
	require.ErrorContains(t, policy.Validate(), "subjective slashing")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DoubleSignSlashBps = 499
	require.ErrorContains(t, policy.Validate(), "double_sign_slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DoubleSignSlashBps = 1_001
	require.ErrorContains(t, policy.Validate(), "double_sign_slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DoubleSignPermanentTombstone = false
	require.ErrorContains(t, policy.Validate(), "permanently tombstone")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ConsensusKeyReuseForbidden = false
	require.ErrorContains(t, policy.Validate(), "consensus key reuse")
}

func TestAetraSlashingParamsSetObjectiveDoubleSignAndBaseDowntime(t *testing.T) {
	params := AetraSlashingParams()

	require.NoError(t, params.Validate())
	require.Equal(t, BpsToLegacyDec(DoubleSignSlashDefaultBps), params.SlashFractionDoubleSign)
	require.Equal(t, BpsToLegacyDec(DowntimeFirstSlashDefaultBps), params.SlashFractionDowntime)
	require.True(t, params.SlashFractionDowntime.GTE(BpsToLegacyDec(DowntimeFirstSlashMinBps)))
	require.True(t, params.SlashFractionDowntime.LTE(BpsToLegacyDec(DowntimeFirstSlashMaxBps)))
	require.Equal(t, time.Duration(DowntimeFirstJailDefaultMinutes)*time.Minute, params.DowntimeJailDuration)
}

func TestSlashingAccountabilityPolicyRejectsUnsafeDowntimeProgression(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.DowntimeFirstSlashBps = 10
	policy.DowntimeRepeatSlashBps = 50
	require.NoError(t, policy.Validate())

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ProgressiveDowntimeEnabled = false
	require.ErrorContains(t, policy.Validate(), "progressive downtime")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.StandardDowntimeStatePreserved = false
	require.ErrorContains(t, policy.Validate(), "x/slashing signing state")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.CustomDowntimeOverlayRequired = false
	require.ErrorContains(t, policy.Validate(), "custom overlay")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeFirstSlashBps = 4
	require.ErrorContains(t, policy.Validate(), "downtime_first_slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeRepeatSlashBps = 51
	require.ErrorContains(t, policy.Validate(), "downtime_repeat_slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeChronicSlashBps = policy.DowntimeRepeatSlashBps
	require.ErrorContains(t, policy.Validate(), "chronic slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeChronicJailMinutes = policy.DowntimeRepeatJailMinutes
	require.ErrorContains(t, policy.Validate(), "chronic jail")
}

func TestSlashingAccountabilityPolicyRejectsUnsafeProposalAndTimestampRules(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalDeterministicReject = false
	require.ErrorContains(t, policy.Validate(), "invalid proposals")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalAutoSlash = true
	require.ErrorContains(t, policy.Validate(), "auto-slash")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalRepeatEvidenceOnly = false
	require.ErrorContains(t, policy.Validate(), "repeated objective evidence")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ProcessProposalExternalInputs = true
	require.ErrorContains(t, policy.Validate(), "external inputs")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TimestampRejectOutsideBounds = false
	require.ErrorContains(t, policy.Validate(), "outside consensus/application bounds")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TimestampCometBFTCompatible = false
	require.ErrorContains(t, policy.Validate(), "CometBFT-compatible")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TimestampCustomWallClockLogic = true
	require.ErrorContains(t, policy.Validate(), "wall-clock")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TimestampSlashObjectiveEvidenceOnly = false
	require.ErrorContains(t, policy.Validate(), "objective reproducible signed evidence")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TimestampMaxForwardDriftSeconds = 0
	require.ErrorContains(t, policy.Validate(), "forward drift")
}
