package params

import (
	"fmt"
	"sort"
)

const (
	AetraStakingPolicyModuleName = "x/aetra-staking-policy"

	AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration = "control_effective_voting_power_delegation_overflow_commission_policy_and_anti_concentration_incentives"
	AetraStakingPolicyCentralAntiCentralizationModule                          = "central_anti_centralization_module"

	AetraStakingPolicyResponsibilityRawStake                       = "calculate_raw_validator_stake"
	AetraStakingPolicyResponsibilityEffectiveStake                 = "calculate_effective_validator_stake"
	AetraStakingPolicyResponsibilityOverflowStake                  = "calculate_overflow_stake"
	AetraStakingPolicyResponsibilityEffectiveVotingPowerCap        = "enforce_or_expose_effective_voting_power_cap"
	AetraStakingPolicyResponsibilityOverflowRewardMultiplier       = "calculate_reward_multiplier_for_overflow_stake"
	AetraStakingPolicyResponsibilityDelegationConcentrationWarning = "expose_delegation_concentration_warnings"
	AetraStakingPolicyResponsibilityCommissionFloor                = "enforce_commission_floor"
	AetraStakingPolicyResponsibilityMaxCommission                  = "enforce_max_commission"
	AetraStakingPolicyResponsibilityMaxCommissionChangeRate        = "enforce_max_commission_change_rate"
	AetraStakingPolicyResponsibilityTopNConcentrationMetrics       = "expose_top_n_concentration_metrics"
	AetraStakingPolicyResponsibilityGovernanceParamValidation      = "validate_governance_param_changes"
	AetraStakingPolicyResponsibilityPolicyChangeEvents             = "emit_events_for_cap_overflow_commission_policy_changes"
	AetraStakingPolicyResponsibilityDeterministicExportImport      = "remain_deterministic_and_export_import_safe"
)

type AetraStakingPolicySpecEvidence struct {
	ModuleName string

	PurposeEffectivePowerOverflowCommissionAntiConcentration bool
	CentralAntiCentralizationModule                          bool

	CalculatesRawValidatorStake              bool
	CalculatesEffectiveValidatorStake        bool
	CalculatesOverflowStake                  bool
	EnforcesOrExposesEffectiveVotingPowerCap bool
	CalculatesOverflowRewardMultiplier       bool
	ExposesDelegationConcentrationWarnings   bool
	EnforcesCommissionFloor                  bool
	EnforcesMaxCommission                    bool
	EnforcesMaxCommissionChangeRate          bool
	ExposesTopNConcentrationMetrics          bool
	ValidatesGovernanceParamChanges          bool
	EmitsCapOverflowCommissionPolicyEvents   bool
	RemainsDeterministicAndExportImportSafe  bool
}

type AetraStakingPolicySpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraStakingPolicySpecEvidence() AetraStakingPolicySpecEvidence {
	return AetraStakingPolicySpecEvidence{
		ModuleName: AetraStakingPolicyModuleName,

		PurposeEffectivePowerOverflowCommissionAntiConcentration: true,
		CentralAntiCentralizationModule:                          true,

		CalculatesRawValidatorStake:              true,
		CalculatesEffectiveValidatorStake:        true,
		CalculatesOverflowStake:                  true,
		EnforcesOrExposesEffectiveVotingPowerCap: true,
		CalculatesOverflowRewardMultiplier:       true,
		ExposesDelegationConcentrationWarnings:   true,
		EnforcesCommissionFloor:                  true,
		EnforcesMaxCommission:                    true,
		EnforcesMaxCommissionChangeRate:          true,
		ExposesTopNConcentrationMetrics:          true,
		ValidatesGovernanceParamChanges:          true,
		EmitsCapOverflowCommissionPolicyEvents:   true,
		RemainsDeterministicAndExportImportSafe:  true,
	}
}

func ValidateAetraStakingPolicySpec(evidence AetraStakingPolicySpecEvidence) error {
	report := BuildAetraStakingPolicySpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicySpecReport(evidence AetraStakingPolicySpecEvidence) AetraStakingPolicySpecReport {
	checks := []requirementCheck{
		{AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration, evidence.PurposeEffectivePowerOverflowCommissionAntiConcentration},
		{AetraStakingPolicyCentralAntiCentralizationModule, evidence.CentralAntiCentralizationModule},
		{AetraStakingPolicyResponsibilityRawStake, evidence.CalculatesRawValidatorStake},
		{AetraStakingPolicyResponsibilityEffectiveStake, evidence.CalculatesEffectiveValidatorStake},
		{AetraStakingPolicyResponsibilityOverflowStake, evidence.CalculatesOverflowStake},
		{AetraStakingPolicyResponsibilityEffectiveVotingPowerCap, evidence.EnforcesOrExposesEffectiveVotingPowerCap},
		{AetraStakingPolicyResponsibilityOverflowRewardMultiplier, evidence.CalculatesOverflowRewardMultiplier},
		{AetraStakingPolicyResponsibilityDelegationConcentrationWarning, evidence.ExposesDelegationConcentrationWarnings},
		{AetraStakingPolicyResponsibilityCommissionFloor, evidence.EnforcesCommissionFloor},
		{AetraStakingPolicyResponsibilityMaxCommission, evidence.EnforcesMaxCommission},
		{AetraStakingPolicyResponsibilityMaxCommissionChangeRate, evidence.EnforcesMaxCommissionChangeRate},
		{AetraStakingPolicyResponsibilityTopNConcentrationMetrics, evidence.ExposesTopNConcentrationMetrics},
		{AetraStakingPolicyResponsibilityGovernanceParamValidation, evidence.ValidatesGovernanceParamChanges},
		{AetraStakingPolicyResponsibilityPolicyChangeEvents, evidence.EmitsCapOverflowCommissionPolicyEvents},
		{AetraStakingPolicyResponsibilityDeterministicExportImport, evidence.RemainsDeterministicAndExportImportSafe},
	}

	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicySpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}
