package params

import (
	"testing"

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
}

func TestSlashingAccountabilityPolicyRejectsSubjectiveOrWeakDoubleSignRules(t *testing.T) {
	policy := DefaultSlashingAccountabilityPolicy()
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
