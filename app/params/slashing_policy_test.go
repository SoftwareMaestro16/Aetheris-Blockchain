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
	require.True(t, policy.BaseFaultsUseCometBFTEvidence)
	require.True(t, policy.StandardDoubleSignIntegrated)
	require.True(t, policy.StandardLivenessDowntimeIntegrated)
	require.True(t, policy.StandardTombstoneIntegrated)
	require.True(t, policy.StandardJailUnjailIntegrated)
	require.True(t, policy.CustomLogicWrapsStandardOnly)
	require.True(t, policy.CoreSlashingForkForbidden)
	require.True(t, policy.ProgressiveDowntimeEnabled)
	require.True(t, policy.StandardDowntimeStatePreserved)
	require.True(t, policy.CustomDowntimeOverlayRequired)
	require.True(t, policy.DowntimeOffenseTracksValidatorConsAddr)
	require.True(t, policy.DowntimeOffenseTracksOffenseCount)
	require.True(t, policy.DowntimeOffenseTracksFirstOffenseTime)
	require.True(t, policy.DowntimeOffenseTracksLastOffenseTime)
	require.True(t, policy.DowntimeOffenseTracksLastSlashFraction)
	require.True(t, policy.DowntimeOffenseTracksCurrentJail)
	require.Equal(t, int64(30), policy.DowntimeOffenseCleanDecayDays)
	require.Equal(t, int64(100), policy.DowntimeOffenseMaxPenaltyBps)
	require.Equal(t, int64(72*60), policy.DowntimeOffenseMaxJailMinutes)
	require.True(t, policy.DowntimeOffenseDelegatorRiskInherited)
	require.True(t, policy.DowntimeOffenseQueryStatusEnabled)
	require.True(t, policy.DowntimeOffenseUnjailKeepsHistory)
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
	require.True(t, policy.HeightConsensusControlled)
	require.True(t, policy.SingleValidatorHeightControlForbidden)
	require.True(t, policy.SameHeightDoubleSignCovered)
	require.True(t, policy.EquivocationCovered)
	require.True(t, policy.InvalidProposalHeightChecked)
	require.True(t, policy.NonDeterministicAppValidationForbidden)
	require.True(t, policy.EvidenceExpirationChecked)
	require.True(t, policy.UnbondingEvidenceTimingChecked)
	require.Equal(t, uint64(100_000), policy.HeightEvidenceMaxAgeBlocks)
	require.Equal(t, uint64(30_000), policy.HeightUnbondingEvidenceWindowBlocks)
	require.True(t, policy.EvidenceWhileValidatorBondedTest)
	require.True(t, policy.EvidenceWhileValidatorUnbondingTest)
	require.True(t, policy.EvidenceAfterUnbondingBeforeExpiryTest)
	require.True(t, policy.EvidenceAfterExpirationRejectedTest)
	require.True(t, policy.DelegatorInfractionHeightSlashTest)
	require.True(t, policy.TombstoneCapBehaviorTest)
	require.True(t, policy.InvalidTxProposalRejectedTest)
	require.True(t, policy.OversizedProposalRejectedTest)
	require.True(t, policy.MalformedSpecialTxRejectedTest)
	require.True(t, policy.ValidProposalAcceptedTest)
	require.True(t, policy.AllValidatorsProposalDeterminismTest)
	require.True(t, policy.InvalidProposalWallClockForbidden)
	require.True(t, policy.InvalidProposalExternalAPIsForbidden)
	require.True(t, policy.ProcessProposalFragilityForbidden)
	require.Equal(t, uint64(2*1024*1024), policy.InvalidProposalMaxBytes)
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

func TestSlashingAccountabilityPolicyRequiresStandardCosmosIntegration(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.UsesCosmosSlashingAndEvidence = false
	require.ErrorContains(t, policy.Validate(), "Cosmos SDK slashing and evidence")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.BaseFaultsUseCometBFTEvidence = false
	require.ErrorContains(t, policy.Validate(), "CometBFT evidence")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.StandardDoubleSignIntegrated = false
	require.ErrorContains(t, policy.Validate(), "double-sign")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.StandardLivenessDowntimeIntegrated = false
	require.ErrorContains(t, policy.Validate(), "liveness/downtime")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.StandardTombstoneIntegrated = false
	require.ErrorContains(t, policy.Validate(), "tombstone")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.StandardJailUnjailIntegrated = false
	require.ErrorContains(t, policy.Validate(), "jail/unjail")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.CustomLogicWrapsStandardOnly = false
	require.ErrorContains(t, policy.Validate(), "wrap or extend")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.CoreSlashingForkForbidden = false
	require.ErrorContains(t, policy.Validate(), "must not be forked")
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

func TestSlashingAccountabilityPolicyRequiresDowntimeOffenseStateAndRules(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksValidatorConsAddr = false
	require.ErrorContains(t, policy.Validate(), "ValidatorConsAddr")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksOffenseCount = false
	require.ErrorContains(t, policy.Validate(), "OffenseCount")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksFirstOffenseTime = false
	require.ErrorContains(t, policy.Validate(), "FirstOffenseTime")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksLastOffenseTime = false
	require.ErrorContains(t, policy.Validate(), "LastOffenseTime")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksLastSlashFraction = false
	require.ErrorContains(t, policy.Validate(), "LastSlashFraction")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseTracksCurrentJail = false
	require.ErrorContains(t, policy.Validate(), "CurrentJailDuration")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseCleanDecayDays = 0
	require.ErrorContains(t, policy.Validate(), "clean period")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseMaxPenaltyBps = DowntimeChronicSlashMaxBps + 1
	require.ErrorContains(t, policy.Validate(), "maximum penalty")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseMaxJailMinutes = policy.DowntimeRepeatJailMinutes - 1
	require.ErrorContains(t, policy.Validate(), "maximum jail")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseDelegatorRiskInherited = false
	require.ErrorContains(t, policy.Validate(), "delegators")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseQueryStatusEnabled = false
	require.ErrorContains(t, policy.Validate(), "status query")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DowntimeOffenseUnjailKeepsHistory = false
	require.ErrorContains(t, policy.Validate(), "unjail")
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

func TestSlashingAccountabilityPolicyRejectsUnsafeHeightManipulationRules(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.HeightConsensusControlled = false
	require.ErrorContains(t, policy.Validate(), "consensus-controlled")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.SingleValidatorHeightControlForbidden = false
	require.ErrorContains(t, policy.Validate(), "single validator")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.SameHeightDoubleSignCovered = false
	require.ErrorContains(t, policy.Validate(), "same-height double-sign")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EquivocationCovered = false
	require.ErrorContains(t, policy.Validate(), "equivocation")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalHeightChecked = false
	require.ErrorContains(t, policy.Validate(), "proposal height")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.NonDeterministicAppValidationForbidden = false
	require.ErrorContains(t, policy.Validate(), "non-deterministic")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EvidenceExpirationChecked = false
	require.ErrorContains(t, policy.Validate(), "expiration")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.UnbondingEvidenceTimingChecked = false
	require.ErrorContains(t, policy.Validate(), "unbonding")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.HeightEvidenceMaxAgeBlocks = 0
	require.ErrorContains(t, policy.Validate(), "max age")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.HeightUnbondingEvidenceWindowBlocks = 0
	require.ErrorContains(t, policy.Validate(), "unbonding evidence window")
}

func TestSlashingAccountabilityPolicyRequiresEvidenceAndUnbondingCoverage(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.HeightUnbondingEvidenceWindowBlocks = policy.HeightEvidenceMaxAgeBlocks + 1
	require.ErrorContains(t, policy.Validate(), "must not exceed evidence max age")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EvidenceWhileValidatorBondedTest = false
	require.ErrorContains(t, policy.Validate(), "validator bonded")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EvidenceWhileValidatorUnbondingTest = false
	require.ErrorContains(t, policy.Validate(), "validator unbonding")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EvidenceAfterUnbondingBeforeExpiryTest = false
	require.ErrorContains(t, policy.Validate(), "before evidence expiration")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.EvidenceAfterExpirationRejectedTest = false
	require.ErrorContains(t, policy.Validate(), "after expiration")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.DelegatorInfractionHeightSlashTest = false
	require.ErrorContains(t, policy.Validate(), "delegators bonded at infraction height")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.TombstoneCapBehaviorTest = false
	require.ErrorContains(t, policy.Validate(), "tombstone cap")
}

func TestSlashingAccountabilityPolicyRequiresInvalidProposalCoverage(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
	policy.InvalidTxProposalRejectedTest = false
	require.ErrorContains(t, policy.Validate(), "invalid tx proposal")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.OversizedProposalRejectedTest = false
	require.ErrorContains(t, policy.Validate(), "oversized proposal")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.MalformedSpecialTxRejectedTest = false
	require.ErrorContains(t, policy.Validate(), "malformed special tx")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ValidProposalAcceptedTest = false
	require.ErrorContains(t, policy.Validate(), "valid proposal")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.AllValidatorsProposalDeterminismTest = false
	require.ErrorContains(t, policy.Validate(), "across validators")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalWallClockForbidden = false
	require.ErrorContains(t, policy.Validate(), "local wall clock")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalExternalAPIsForbidden = false
	require.ErrorContains(t, policy.Validate(), "external APIs")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.ProcessProposalFragilityForbidden = false
	require.ErrorContains(t, policy.Validate(), "not be fragile")

	policy = DefaultSlashingAccountabilityPolicy()
	policy.InvalidProposalMaxBytes = 0
	require.ErrorContains(t, policy.Validate(), "max bytes")
}
