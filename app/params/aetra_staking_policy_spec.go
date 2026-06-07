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

	AetraStakingPolicyStateParams                = "Params"
	AetraStakingPolicyStateValidatorPolicy       = "ValidatorPolicy"
	AetraStakingPolicyStateConcentrationSnapshot = "ConcentrationSnapshot"

	AetraStakingPolicyStateParamMaxValidatorsSoftTarget        = "MaxValidatorsSoftTarget"
	AetraStakingPolicyStateParamValidatorPowerCapBps           = "ValidatorPowerCapBps"
	AetraStakingPolicyStateParamValidatorPowerCapSchedule      = "ValidatorPowerCapSchedule"
	AetraStakingPolicyStateParamOverflowRewardMultiplierBps    = "OverflowRewardMultiplierBps"
	AetraStakingPolicyStateParamCommissionFloorBps             = "CommissionFloorBps"
	AetraStakingPolicyStateParamCommissionMaxBps               = "CommissionMaxBps"
	AetraStakingPolicyStateParamCommissionMaxDailyChangeBps    = "CommissionMaxDailyChangeBps"
	AetraStakingPolicyStateParamTop10TargetBps                 = "Top10TargetBps"
	AetraStakingPolicyStateParamTop20TargetBps                 = "Top20TargetBps"
	AetraStakingPolicyStateParamTop33TargetBps                 = "Top33TargetBps"
	AetraStakingPolicyStateParamMinSelfBond                    = "MinSelfBond"
	AetraStakingPolicyStateParamMinValidatorBond               = "MinValidatorBond"
	AetraStakingPolicyStateParamWarningThresholdBps            = "WarningThresholdBps"
	AetraStakingPolicyStateValidatorOperatorAddress            = "OperatorAddress"
	AetraStakingPolicyStateValidatorRawBondedTokens            = "RawBondedTokens"
	AetraStakingPolicyStateValidatorEffectiveBondedTokens      = "EffectiveBondedTokens"
	AetraStakingPolicyStateValidatorOverflowBondedTokens       = "OverflowBondedTokens"
	AetraStakingPolicyStateValidatorEffectivePowerBps          = "EffectivePowerBps"
	AetraStakingPolicyStateValidatorIsOverCap                  = "IsOverCap"
	AetraStakingPolicyStateValidatorRewardMultiplierBps        = "RewardMultiplierBps"
	AetraStakingPolicyStateValidatorLastCommissionChangeTime   = "LastCommissionChangeTime"
	AetraStakingPolicyStateValidatorLastCommissionRateBps      = "LastCommissionRateBps"
	AetraStakingPolicyStateSnapshotHeight                      = "Height"
	AetraStakingPolicyStateSnapshotBondedRatio                 = "BondedRatio"
	AetraStakingPolicyStateSnapshotActiveValidators            = "ActiveValidators"
	AetraStakingPolicyStateSnapshotTop10Bps                    = "Top10Bps"
	AetraStakingPolicyStateSnapshotTop20Bps                    = "Top20Bps"
	AetraStakingPolicyStateSnapshotTop33Bps                    = "Top33Bps"
	AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate = "NakamotoCoefficientEstimate"
	AetraStakingPolicyStateIntegerBpsOrSDKDecimal              = "integer_basis_points_or_sdk_decimal_accounting"
	AetraStakingPolicyStateNoFloatingPoint                     = "avoid_floating_point_accounting"
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

type AetraStakingPolicyStateSpecEvidence struct {
	ModuleName string

	ParamsFields                []string
	ValidatorPolicyFields       []string
	ConcentrationSnapshotFields []string

	IntegerBasisPointsOrSDKDecimals bool
	AvoidsFloatingPoint             bool
}

type AetraStakingPolicyStateSpecReport struct {
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

func DefaultAetraStakingPolicyStateSpecEvidence() AetraStakingPolicyStateSpecEvidence {
	return AetraStakingPolicyStateSpecEvidence{
		ModuleName: AetraStakingPolicyModuleName,
		ParamsFields: []string{
			AetraStakingPolicyStateParamMaxValidatorsSoftTarget,
			AetraStakingPolicyStateParamValidatorPowerCapBps,
			AetraStakingPolicyStateParamValidatorPowerCapSchedule,
			AetraStakingPolicyStateParamOverflowRewardMultiplierBps,
			AetraStakingPolicyStateParamCommissionFloorBps,
			AetraStakingPolicyStateParamCommissionMaxBps,
			AetraStakingPolicyStateParamCommissionMaxDailyChangeBps,
			AetraStakingPolicyStateParamTop10TargetBps,
			AetraStakingPolicyStateParamTop20TargetBps,
			AetraStakingPolicyStateParamTop33TargetBps,
			AetraStakingPolicyStateParamMinSelfBond,
			AetraStakingPolicyStateParamMinValidatorBond,
			AetraStakingPolicyStateParamWarningThresholdBps,
		},
		ValidatorPolicyFields: []string{
			AetraStakingPolicyStateValidatorOperatorAddress,
			AetraStakingPolicyStateValidatorRawBondedTokens,
			AetraStakingPolicyStateValidatorEffectiveBondedTokens,
			AetraStakingPolicyStateValidatorOverflowBondedTokens,
			AetraStakingPolicyStateValidatorEffectivePowerBps,
			AetraStakingPolicyStateValidatorIsOverCap,
			AetraStakingPolicyStateValidatorRewardMultiplierBps,
			AetraStakingPolicyStateValidatorLastCommissionChangeTime,
			AetraStakingPolicyStateValidatorLastCommissionRateBps,
		},
		ConcentrationSnapshotFields: []string{
			AetraStakingPolicyStateSnapshotHeight,
			AetraStakingPolicyStateSnapshotBondedRatio,
			AetraStakingPolicyStateSnapshotActiveValidators,
			AetraStakingPolicyStateSnapshotTop10Bps,
			AetraStakingPolicyStateSnapshotTop20Bps,
			AetraStakingPolicyStateSnapshotTop33Bps,
			AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate,
		},
		IntegerBasisPointsOrSDKDecimals: true,
		AvoidsFloatingPoint:             true,
	}
}

func ValidateAetraStakingPolicyStateSpec(evidence AetraStakingPolicyStateSpecEvidence) error {
	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra staking policy state spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraStakingPolicyStateSpecReport(evidence AetraStakingPolicyStateSpecEvidence) AetraStakingPolicyStateSpecReport {
	requiredParams := requiredAetraStakingPolicyStateParamsFields()
	requiredValidator := requiredAetraStakingPolicyStateValidatorPolicyFields()
	requiredSnapshot := requiredAetraStakingPolicyStateConcentrationSnapshotFields()

	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraStakingPolicyModuleName {
		failed = append(failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	}

	passed := 0
	paramsPassed, paramsFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateParams, evidence.ParamsFields, requiredParams)
	passed += paramsPassed
	failed = append(failed, paramsFailed...)

	validatorPassed, validatorFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateValidatorPolicy, evidence.ValidatorPolicyFields, requiredValidator)
	passed += validatorPassed
	failed = append(failed, validatorFailed...)

	snapshotPassed, snapshotFailed := validateAetraStakingPolicyStateFields(AetraStakingPolicyStateConcentrationSnapshot, evidence.ConcentrationSnapshotFields, requiredSnapshot)
	passed += snapshotPassed
	failed = append(failed, snapshotFailed...)

	for _, check := range []requirementCheck{
		{AetraStakingPolicyStateIntegerBpsOrSDKDecimal, evidence.IntegerBasisPointsOrSDKDecimals},
		{AetraStakingPolicyStateNoFloatingPoint, evidence.AvoidsFloatingPoint},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraStakingPolicyStateSpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(requiredParams) + len(requiredValidator) + len(requiredSnapshot) + 2,
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func requiredAetraStakingPolicyStateParamsFields() []string {
	return []string{
		AetraStakingPolicyStateParamMaxValidatorsSoftTarget,
		AetraStakingPolicyStateParamValidatorPowerCapBps,
		AetraStakingPolicyStateParamValidatorPowerCapSchedule,
		AetraStakingPolicyStateParamOverflowRewardMultiplierBps,
		AetraStakingPolicyStateParamCommissionFloorBps,
		AetraStakingPolicyStateParamCommissionMaxBps,
		AetraStakingPolicyStateParamCommissionMaxDailyChangeBps,
		AetraStakingPolicyStateParamTop10TargetBps,
		AetraStakingPolicyStateParamTop20TargetBps,
		AetraStakingPolicyStateParamTop33TargetBps,
		AetraStakingPolicyStateParamMinSelfBond,
		AetraStakingPolicyStateParamMinValidatorBond,
		AetraStakingPolicyStateParamWarningThresholdBps,
	}
}

func requiredAetraStakingPolicyStateValidatorPolicyFields() []string {
	return []string{
		AetraStakingPolicyStateValidatorOperatorAddress,
		AetraStakingPolicyStateValidatorRawBondedTokens,
		AetraStakingPolicyStateValidatorEffectiveBondedTokens,
		AetraStakingPolicyStateValidatorOverflowBondedTokens,
		AetraStakingPolicyStateValidatorEffectivePowerBps,
		AetraStakingPolicyStateValidatorIsOverCap,
		AetraStakingPolicyStateValidatorRewardMultiplierBps,
		AetraStakingPolicyStateValidatorLastCommissionChangeTime,
		AetraStakingPolicyStateValidatorLastCommissionRateBps,
	}
}

func requiredAetraStakingPolicyStateConcentrationSnapshotFields() []string {
	return []string{
		AetraStakingPolicyStateSnapshotHeight,
		AetraStakingPolicyStateSnapshotBondedRatio,
		AetraStakingPolicyStateSnapshotActiveValidators,
		AetraStakingPolicyStateSnapshotTop10Bps,
		AetraStakingPolicyStateSnapshotTop20Bps,
		AetraStakingPolicyStateSnapshotTop33Bps,
		AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate,
	}
}

func validateAetraStakingPolicyStateFields(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, field := range required {
		requiredSet[field] = true
	}
	seen := map[string]bool{}
	for _, field := range actual {
		if field == "" {
			failed = append(failed, group+".field_name_required")
			continue
		}
		if seen[field] {
			failed = append(failed, group+"."+field+":duplicate")
			continue
		}
		seen[field] = true
		if !requiredSet[field] {
			failed = append(failed, group+"."+field+":unexpected")
		}
	}
	passed := 0
	for _, field := range required {
		if seen[field] {
			passed++
		} else {
			failed = append(failed, group+"."+field+":missing")
		}
	}
	return passed, failed
}
