package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationPhasePlansCoverPhase0AndPhase1(t *testing.T) {
	plans := DefaultImplementationPhasePlans()
	require.Len(t, plans, 2)

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
	plan := DefaultImplementationPhasePlans()[1]
	plan.Items = plan.Items[:len(plan.Items)-1]

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, PhaseAcceptanceDeterministicExportImport+":missing")
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
