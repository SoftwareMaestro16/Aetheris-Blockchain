package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationPhasePlansCoverPhase0ThroughPhase2(t *testing.T) {
	plans := DefaultImplementationPhasePlans()
	require.Len(t, plans, 3)

	for _, plan := range plans {
		report := BuildImplementationPhaseReport(plan)
		require.True(t, report.Ready, report.Failed)
		require.Empty(t, report.Failed)
		require.Equal(t, report.Required, report.Done)
		require.NoError(t, ValidateImplementationPhasePlan(plan))
	}
}

func TestImplementationPhaseRejectsMissingEvidence(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[0]
	plan.Items[0].Done = false

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, PhaseTaskInspectVersions+":missing_evidence")
	require.Error(t, ValidateImplementationPhasePlan(plan))
}

func TestImplementationPhaseRejectsMissingRequiredItem(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[2]
	plan.Items = plan.Items[:len(plan.Items)-1]

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, PhaseAcceptanceRewardsDeterministic+":missing")
}

func TestImplementationPhaseEconomicsFeeSplitRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[2]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
		PhaseTaskImplementInflationBounds,
		PhaseTaskImplementTargetBondedRatio,
		PhaseTaskImplementFeeSplit,
		PhaseTaskImplementRewardSmoothing,
		PhaseTaskExposeAPREstimateQuery,
		PhaseTaskExposeSupplyTreasuryQueries,
		PhaseTaskAddEconomicsGovernanceParams,
		PhaseTestInflationCurve,
		PhaseTestBondedRatio,
		PhaseTestFeeSplit,
		PhaseTestBurnAccounting,
		PhaseTestTreasuryAccounting,
		PhaseTestAPRQuery,
		PhaseTestSupplyInvariant,
		PhaseTestEconomicsExportImport,
		PhaseAcceptanceInflationWithinBounds,
		PhaseAcceptanceFeeSplitSumsToFullAmount,
		PhaseAcceptanceBurnReducesSupply,
		PhaseAcceptanceTreasuryReceivesAmount,
		PhaseAcceptanceRewardsDeterministic,
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseRejectsUnknownPhaseAndUnexpectedItem(t *testing.T) {
	plan := ImplementationPhasePlan{
		PhaseID: "phase_99",
		Items: []ImplementationPhaseItem{
			phaseItem("task", PhaseTaskInspectVersions),
		},
	}

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "phase_99:unknown_phase")
	require.Contains(t, report.Failed, PhaseTaskInspectVersions+":unexpected")
}
