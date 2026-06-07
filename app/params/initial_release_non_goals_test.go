package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultInitialReleaseScopeRespectsNonGoals(t *testing.T) {
	policy := DefaultInitialReleaseScopePolicy()

	report := BuildInitialReleaseNonGoalsReport(policy)
	require.True(t, report.Allowed, report.Violations)
	require.Empty(t, report.Violations)
	require.Equal(t, 9, report.NonGoalsChecked)
	require.NoError(t, ValidateInitialReleaseScope(policy))
}

func TestInitialReleaseScopeRejectsPerformanceAndValidatorScopeCreep(t *testing.T) {
	policy := DefaultInitialReleaseScopePolicy()
	policy.AttemptsPoH = true
	policy.AttemptsSolanaLevelTPS = true
	policy.AllowsOneSecondBlocks = true
	policy.RequiresMandatoryKYCValidators = true
	policy.AllowsUnlimitedValidatorSet = true

	report := BuildInitialReleaseNonGoalsReport(policy)
	require.False(t, report.Allowed)
	require.Contains(t, report.Violations, InitialReleaseNonGoalPoH)
	require.Contains(t, report.Violations, InitialReleaseNonGoalSolanaLevelTPS)
	require.Contains(t, report.Violations, InitialReleaseNonGoalOneSecondBlocks)
	require.Contains(t, report.Violations, InitialReleaseNonGoalMandatoryKYC)
	require.Contains(t, report.Violations, InitialReleaseNonGoalUnlimitedValidatorSet)
	require.Error(t, ValidateInitialReleaseScope(policy))
}

func TestInitialReleaseScopeRejectsExecutionSecurityAndMarketingScopeCreep(t *testing.T) {
	policy := DefaultInitialReleaseScopePolicy()
	policy.EnablesEVMAtGenesisWithoutApproval = true
	policy.AllowsSubjectiveSlashing = true
	policy.AllowsUnboundedContractExecution = true
	policy.UsesHighInflationAPRMarketing = true

	report := BuildInitialReleaseNonGoalsReport(policy)
	require.False(t, report.Allowed)
	require.Contains(t, report.Violations, InitialReleaseNonGoalEVMAtGenesis)
	require.Contains(t, report.Violations, InitialReleaseNonGoalSubjectiveSlashing)
	require.Contains(t, report.Violations, InitialReleaseNonGoalUnboundedContractExecution)
	require.Contains(t, report.Violations, InitialReleaseNonGoalHighInflationAPRMarketing)
	require.Error(t, ValidateInitialReleaseScope(policy))
}
