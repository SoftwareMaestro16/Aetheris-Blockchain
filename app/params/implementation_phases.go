package params

import (
	"fmt"
	"sort"
)

const (
	ImplementationPhaseBaselineAudit    = "phase_0_baseline_audit"
	ImplementationPhaseStakingPolicyCap = "phase_1_staking_policy_validator_cap"

	PhaseTaskInspectVersions             = "inspect_current_cosmos_sdk_and_cometbft_versions"
	PhaseTaskDocumentModuleGraph         = "document_current_app_module_graph"
	PhaseTaskIdentifyOverlappingModules  = "identify_modules_overlapping_custom_aetra_modules"
	PhaseTaskDecideRenameReuseWrap       = "decide_modules_renamed_reused_or_wrapped"
	PhaseTaskVerifyNaetStakingDenom      = "verify_naet_staking_denom"
	PhaseTaskVerifyEconomyWiring         = "verify_fee_collector_burn_treasury_emissions_mint_authority_wiring"
	PhaseTaskVerifyLocalnetAndCoverage   = "verify_localnet_scripts_and_test_coverage"
	PhaseTaskImplementEffectivePowerCap  = "implement_effective_voting_power_cap"
	PhaseTaskImplementOverflowAccounting = "implement_overflow_stake_accounting"
	PhaseTaskImplementCommissionPolicy   = "implement_commission_floor_max_change_policy"
	PhaseTaskAddConcentrationMetrics     = "add_concentration_metrics"
	PhaseTaskAddStakeQueries             = "add_validator_raw_effective_overflow_queries"
	PhaseTaskAddGovernanceParams         = "add_governance_params_with_validation"
	PhaseTaskWireModuleLifecycle         = "wire_module_into_app_lifecycle"

	PhaseDeliverableModuleInventory         = "module_inventory"
	PhaseDeliverableGapAnalysis             = "gap_analysis"
	PhaseDeliverableRiskList                = "risk_list"
	PhaseDeliverableImplementationChecklist = "updated_implementation_checklist"

	PhaseTestFullUnitRun            = "current_full_unit_test_run"
	PhaseTestIntegrationRun         = "current_integration_test_run"
	PhaseTestLocalnetSmoke          = "current_localnet_smoke_test"
	PhaseTestExportImport           = "current_export_import_test"
	PhaseTestCapMathUnit            = "cap_math_unit_tests"
	PhaseTestValidatorSetTransition = "validator_set_transition_tests"
	PhaseTestConcentrationQuery     = "concentration_query_tests"
	PhaseTestCommissionBounds       = "commission_bounds_tests"
	PhaseTestStakingIntegration     = "integration_tests_with_staking"
	PhaseTestStakingExportImport    = "staking_policy_export_import_tests"
	PhaseTestInvariant              = "invariant_tests"

	PhaseAcceptanceNoValidatorExceedsCap     = "no_validator_can_exceed_effective_power_cap"
	PhaseAcceptanceExcessNoVotingPower       = "excess_stake_does_not_increase_voting_power"
	PhaseAcceptanceParamsSafeBounds          = "params_cannot_be_set_outside_safe_bounds"
	PhaseAcceptanceDeterministicExportImport = "state_remains_deterministic_after_export_import"
)

type ImplementationPhaseItem struct {
	ID       string
	Kind     string
	Required bool
	Done     bool
	Evidence string
}

type ImplementationPhasePlan struct {
	PhaseID string
	Items   []ImplementationPhaseItem
}

type ImplementationPhaseReport struct {
	PhaseID  string
	Required int
	Done     int
	Failed   []string
	Ready    bool
}

func DefaultImplementationPhasePlans() []ImplementationPhasePlan {
	return []ImplementationPhasePlan{
		{
			PhaseID: ImplementationPhaseBaselineAudit,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskInspectVersions),
				phaseItem("task", PhaseTaskDocumentModuleGraph),
				phaseItem("task", PhaseTaskIdentifyOverlappingModules),
				phaseItem("task", PhaseTaskDecideRenameReuseWrap),
				phaseItem("task", PhaseTaskVerifyNaetStakingDenom),
				phaseItem("task", PhaseTaskVerifyEconomyWiring),
				phaseItem("task", PhaseTaskVerifyLocalnetAndCoverage),
				phaseItem("deliverable", PhaseDeliverableModuleInventory),
				phaseItem("deliverable", PhaseDeliverableGapAnalysis),
				phaseItem("deliverable", PhaseDeliverableRiskList),
				phaseItem("deliverable", PhaseDeliverableImplementationChecklist),
				phaseItem("test", PhaseTestFullUnitRun),
				phaseItem("test", PhaseTestIntegrationRun),
				phaseItem("test", PhaseTestLocalnetSmoke),
				phaseItem("test", PhaseTestExportImport),
			},
		},
		{
			PhaseID: ImplementationPhaseStakingPolicyCap,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskImplementEffectivePowerCap),
				phaseItem("task", PhaseTaskImplementOverflowAccounting),
				phaseItem("task", PhaseTaskImplementCommissionPolicy),
				phaseItem("task", PhaseTaskAddConcentrationMetrics),
				phaseItem("task", PhaseTaskAddStakeQueries),
				phaseItem("task", PhaseTaskAddGovernanceParams),
				phaseItem("task", PhaseTaskWireModuleLifecycle),
				phaseItem("test", PhaseTestCapMathUnit),
				phaseItem("test", PhaseTestValidatorSetTransition),
				phaseItem("test", PhaseTestConcentrationQuery),
				phaseItem("test", PhaseTestCommissionBounds),
				phaseItem("test", PhaseTestStakingIntegration),
				phaseItem("test", PhaseTestStakingExportImport),
				phaseItem("test", PhaseTestInvariant),
				phaseItem("acceptance", PhaseAcceptanceNoValidatorExceedsCap),
				phaseItem("acceptance", PhaseAcceptanceExcessNoVotingPower),
				phaseItem("acceptance", PhaseAcceptanceParamsSafeBounds),
				phaseItem("acceptance", PhaseAcceptanceDeterministicExportImport),
			},
		},
	}
}

func ValidateImplementationPhasePlan(plan ImplementationPhasePlan) error {
	report := BuildImplementationPhaseReport(plan)
	if !report.Ready {
		return fmt.Errorf("implementation phase %s failed: %v", report.PhaseID, report.Failed)
	}
	return nil
}

func BuildImplementationPhaseReport(plan ImplementationPhasePlan) ImplementationPhaseReport {
	expected := expectedImplementationPhaseItems(plan.PhaseID)
	failed := make([]string, 0)
	seen := map[string]ImplementationPhaseItem{}
	required := 0
	done := 0
	if plan.PhaseID == "" {
		failed = append(failed, "phase_id_required")
	}
	if len(expected) == 0 {
		failed = append(failed, plan.PhaseID+":unknown_phase")
	}
	for _, item := range plan.Items {
		if item.ID == "" || item.Kind == "" {
			failed = append(failed, "phase_item_id_and_kind_required")
			continue
		}
		if _, duplicate := seen[item.ID]; duplicate {
			failed = append(failed, item.ID+":duplicate")
		}
		seen[item.ID] = item
		if !expected[item.ID] {
			failed = append(failed, item.ID+":unexpected")
		}
		if item.Required {
			required++
		}
		if item.Required && (!item.Done || item.Evidence == "") {
			failed = append(failed, item.ID+":missing_evidence")
		}
		if item.Required && item.Done && item.Evidence != "" {
			done++
		}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			failed = append(failed, id+":missing")
		}
	}
	sort.Strings(failed)
	return ImplementationPhaseReport{
		PhaseID:  plan.PhaseID,
		Required: required,
		Done:     done,
		Failed:   failed,
		Ready:    len(failed) == 0,
	}
}

func phaseItem(kind, id string) ImplementationPhaseItem {
	return ImplementationPhaseItem{
		ID:       id,
		Kind:     kind,
		Required: true,
		Done:     true,
		Evidence: "required " + kind + " evidence for " + id,
	}
}

func expectedImplementationPhaseItems(phaseID string) map[string]bool {
	out := map[string]bool{}
	for _, plan := range defaultImplementationPhaseItemIDs() {
		if plan.phaseID != phaseID {
			continue
		}
		for _, id := range plan.ids {
			out[id] = true
		}
	}
	return out
}

type phaseItemIDs struct {
	phaseID string
	ids     []string
}

func defaultImplementationPhaseItemIDs() []phaseItemIDs {
	return []phaseItemIDs{
		{
			phaseID: ImplementationPhaseBaselineAudit,
			ids: []string{
				PhaseTaskInspectVersions,
				PhaseTaskDocumentModuleGraph,
				PhaseTaskIdentifyOverlappingModules,
				PhaseTaskDecideRenameReuseWrap,
				PhaseTaskVerifyNaetStakingDenom,
				PhaseTaskVerifyEconomyWiring,
				PhaseTaskVerifyLocalnetAndCoverage,
				PhaseDeliverableModuleInventory,
				PhaseDeliverableGapAnalysis,
				PhaseDeliverableRiskList,
				PhaseDeliverableImplementationChecklist,
				PhaseTestFullUnitRun,
				PhaseTestIntegrationRun,
				PhaseTestLocalnetSmoke,
				PhaseTestExportImport,
			},
		},
		{
			phaseID: ImplementationPhaseStakingPolicyCap,
			ids: []string{
				PhaseTaskImplementEffectivePowerCap,
				PhaseTaskImplementOverflowAccounting,
				PhaseTaskImplementCommissionPolicy,
				PhaseTaskAddConcentrationMetrics,
				PhaseTaskAddStakeQueries,
				PhaseTaskAddGovernanceParams,
				PhaseTaskWireModuleLifecycle,
				PhaseTestCapMathUnit,
				PhaseTestValidatorSetTransition,
				PhaseTestConcentrationQuery,
				PhaseTestCommissionBounds,
				PhaseTestStakingIntegration,
				PhaseTestStakingExportImport,
				PhaseTestInvariant,
				PhaseAcceptanceNoValidatorExceedsCap,
				PhaseAcceptanceExcessNoVotingPower,
				PhaseAcceptanceParamsSafeBounds,
				PhaseAcceptanceDeterministicExportImport,
			},
		},
	}
}
