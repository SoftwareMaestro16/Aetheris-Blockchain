package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraStakingPolicySpecCoversResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraStakingPolicyModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 15, report.Required)
	require.NoError(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingStakeAndPowerResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.CalculatesRawValidatorStake = false
	evidence.CalculatesEffectiveValidatorStake = false
	evidence.CalculatesOverflowStake = false
	evidence.EnforcesOrExposesEffectiveVotingPowerCap = false
	evidence.CalculatesOverflowRewardMultiplier = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityRawStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityEffectiveStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityOverflowStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityEffectiveVotingPowerCap)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityOverflowRewardMultiplier)
	require.Error(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingCommissionConcentrationAndSafetyResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.ExposesDelegationConcentrationWarnings = false
	evidence.EnforcesCommissionFloor = false
	evidence.EnforcesMaxCommission = false
	evidence.EnforcesMaxCommissionChangeRate = false
	evidence.ExposesTopNConcentrationMetrics = false
	evidence.ValidatesGovernanceParamChanges = false
	evidence.EmitsCapOverflowCommissionPolicyEvents = false
	evidence.RemainsDeterministicAndExportImportSafe = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityDelegationConcentrationWarning)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityCommissionFloor)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityMaxCommission)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityMaxCommissionChangeRate)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityTopNConcentrationMetrics)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityGovernanceParamValidation)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityPolicyChangeEvents)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityDeterministicExportImport)
	require.Error(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingPurposeAndCentralModuleRole(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.PurposeEffectivePowerOverflowCommissionAntiConcentration = false
	evidence.CentralAntiCentralizationModule = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration)
	require.Contains(t, report.Failed, AetraStakingPolicyCentralAntiCentralizationModule)
}

func TestAetraStakingPolicySpecRejectsWrongModuleIdentity(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.ModuleName = "x/staking-policy"

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)

	evidence.ModuleName = ""
	report = BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
}
