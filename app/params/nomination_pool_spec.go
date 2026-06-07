package params

import (
	"fmt"
	"sort"
)

const (
	AetraNominationPoolModuleName = "x/nominator-pool"

	AetraNominationPoolPurposeAccessibilityWithAccountingAndCentralizationRisk = "nomination_pools_improve_accessibility_but_introduce_accounting_and_centralization_risks"

	AetraNominationPoolStatePool           = "Pool"
	AetraNominationPoolStatePoolDelegation = "PoolDelegation"

	AetraNominationPoolFieldPoolID           = "PoolId"
	AetraNominationPoolFieldOperatorAddress  = "OperatorAddress"
	AetraNominationPoolFieldValidatorAddress = "ValidatorAddress"
	AetraNominationPoolFieldTotalBonded      = "TotalBonded"
	AetraNominationPoolFieldTotalShares      = "TotalShares"
	AetraNominationPoolFieldCommissionBps    = "CommissionBps"
	AetraNominationPoolFieldStatus           = "Status"
	AetraNominationPoolFieldCreatedHeight    = "CreatedHeight"
	AetraNominationPoolFieldUnbondingEntries = "UnbondingEntries"

	AetraNominationPoolFieldDelegatorAddress  = "DelegatorAddress"
	AetraNominationPoolFieldDelegationPoolID  = "PoolId"
	AetraNominationPoolFieldShares            = "Shares"
	AetraNominationPoolFieldPrincipalEstimate = "PrincipalEstimate"
	AetraNominationPoolFieldRewardsAccrued    = "RewardsAccrued"

	AetraNominationPoolRiskAccessibility  = "accessibility_for_users_without_validator_infrastructure"
	AetraNominationPoolRiskAccounting     = "deterministic_pool_share_principal_reward_accounting"
	AetraNominationPoolRiskCentralization = "pool_operator_and_validator_concentration_risk"
	AetraNominationPoolImplementationMap  = "current_x_nominator_pool_state_mapping_required"
)

type AetraNominationPoolModelEvidence struct {
	ModuleName string

	PoolFields           []string
	PoolDelegationFields []string

	AccessibilityRiskAcknowledged  bool
	AccountingRiskAcknowledged     bool
	CentralizationRiskAcknowledged bool
	CurrentImplementationMapped    bool
}

type AetraNominationPoolModelReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraNominationPoolModelEvidence() AetraNominationPoolModelEvidence {
	return AetraNominationPoolModelEvidence{
		ModuleName: AetraNominationPoolModuleName,
		PoolFields: []string{
			AetraNominationPoolFieldPoolID,
			AetraNominationPoolFieldOperatorAddress,
			AetraNominationPoolFieldValidatorAddress,
			AetraNominationPoolFieldTotalBonded,
			AetraNominationPoolFieldTotalShares,
			AetraNominationPoolFieldCommissionBps,
			AetraNominationPoolFieldStatus,
			AetraNominationPoolFieldCreatedHeight,
			AetraNominationPoolFieldUnbondingEntries,
		},
		PoolDelegationFields: []string{
			AetraNominationPoolFieldDelegatorAddress,
			AetraNominationPoolFieldDelegationPoolID,
			AetraNominationPoolFieldShares,
			AetraNominationPoolFieldPrincipalEstimate,
			AetraNominationPoolFieldRewardsAccrued,
		},
		AccessibilityRiskAcknowledged:  true,
		AccountingRiskAcknowledged:     true,
		CentralizationRiskAcknowledged: true,
		CurrentImplementationMapped:    true,
	}
}

func ValidateAetraNominationPoolModel(evidence AetraNominationPoolModelEvidence) error {
	report := BuildAetraNominationPoolModelReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra nomination pool model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraNominationPoolModelReport(evidence AetraNominationPoolModelEvidence) AetraNominationPoolModelReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraNominationPoolModuleName {
		failed = append(failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	}

	requiredPool := requiredAetraNominationPoolFields()
	requiredDelegation := requiredAetraNominationPoolDelegationFields()
	passedPool, failedPool := validateAetraNominationPoolCatalog(AetraNominationPoolStatePool, evidence.PoolFields, requiredPool)
	passedDelegation, failedDelegation := validateAetraNominationPoolCatalog(AetraNominationPoolStatePoolDelegation, evidence.PoolDelegationFields, requiredDelegation)
	failed = append(failed, failedPool...)
	failed = append(failed, failedDelegation...)

	checks := []requirementCheck{
		{AetraNominationPoolRiskAccessibility, evidence.AccessibilityRiskAcknowledged},
		{AetraNominationPoolRiskAccounting, evidence.AccountingRiskAcknowledged},
		{AetraNominationPoolRiskCentralization, evidence.CentralizationRiskAcknowledged},
		{AetraNominationPoolImplementationMap, evidence.CurrentImplementationMapped},
	}
	passedChecks := 0
	for _, check := range checks {
		if check.Passed {
			passedChecks++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraNominationPoolModelReport{
		ModuleName: evidence.ModuleName,
		Required:   len(requiredPool) + len(requiredDelegation) + len(checks),
		Passed:     passedPool + passedDelegation + passedChecks,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func requiredAetraNominationPoolFields() []string {
	return []string{
		AetraNominationPoolFieldPoolID,
		AetraNominationPoolFieldOperatorAddress,
		AetraNominationPoolFieldValidatorAddress,
		AetraNominationPoolFieldTotalBonded,
		AetraNominationPoolFieldTotalShares,
		AetraNominationPoolFieldCommissionBps,
		AetraNominationPoolFieldStatus,
		AetraNominationPoolFieldCreatedHeight,
		AetraNominationPoolFieldUnbondingEntries,
	}
}

func requiredAetraNominationPoolDelegationFields() []string {
	return []string{
		AetraNominationPoolFieldDelegatorAddress,
		AetraNominationPoolFieldDelegationPoolID,
		AetraNominationPoolFieldShares,
		AetraNominationPoolFieldPrincipalEstimate,
		AetraNominationPoolFieldRewardsAccrued,
	}
}

func validateAetraNominationPoolCatalog(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, item := range required {
		requiredSet[item] = true
	}
	seen := map[string]bool{}
	for _, item := range actual {
		if item == "" {
			failed = append(failed, group+".item_required")
			continue
		}
		if seen[item] {
			failed = append(failed, group+"."+item+":duplicate")
			continue
		}
		seen[item] = true
		if !requiredSet[item] {
			failed = append(failed, group+"."+item+":unexpected")
		}
	}
	passed := 0
	for _, item := range required {
		if seen[item] {
			passed++
		} else {
			failed = append(failed, group+"."+item+":missing")
		}
	}
	return passed, failed
}
