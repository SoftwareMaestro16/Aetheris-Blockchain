package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultFeeSplitPolicyMatchesAetraEconomics(t *testing.T) {
	policy := DefaultFeeSplitPolicy()

	require.NoError(t, policy.Validate())
	require.Equal(t, int64(5_000), policy.BurnBps)
	require.Equal(t, int64(3_500), policy.ValidatorDelegatorBps)
	require.Equal(t, int64(1_500), policy.TreasuryBps)
	require.True(t, policy.GovernanceConfigurable)
}

func TestFeeSplitPolicyRejectsUnsafeGovernanceParams(t *testing.T) {
	policy := DefaultFeeSplitPolicy()
	policy.BurnBps = 2_000
	policy.ValidatorDelegatorBps = 5_000
	require.ErrorContains(t, policy.Validate(), "fee_burn_bps")

	policy = DefaultFeeSplitPolicy()
	policy.GovernanceConfigurable = false
	require.ErrorContains(t, policy.Validate(), "governance configurable")
}

func TestApproximateStakingAPRExamples(t *testing.T) {
	apr, err := ApproximateStakingAPRBps(300, 6_000)
	require.NoError(t, err)
	require.Equal(t, int64(500), apr)

	apr, err = ApproximateStakingAPRBps(400, 6_000)
	require.NoError(t, err)
	require.Equal(t, int64(667), apr)

	apr, err = ApproximateStakingAPRBps(500, 6_000)
	require.NoError(t, err)
	require.Equal(t, int64(833), apr)
}

func TestApproximateStakingAPRRejectsZeroBondedRatio(t *testing.T) {
	_, err := ApproximateStakingAPRBps(300, 0)
	require.ErrorContains(t, err, "bonded_ratio_bps")
}
