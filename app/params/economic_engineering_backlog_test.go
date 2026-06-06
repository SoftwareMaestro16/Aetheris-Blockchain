package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultEconomicEngineeringBacklogIsTrackedByPriority(t *testing.T) {
	report := BuildEconomicEngineeringBacklogReport(nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Items, 23)
	require.Equal(t, 10, report.RequiredHigh)
	require.Equal(t, 7, report.RequiredMedium)
	require.Equal(t, 6, report.RequiredLower)
	require.Equal(t, 10, report.CoveredHigh)
	require.Equal(t, 7, report.CoveredMedium)
	require.Equal(t, 6, report.CoveredLower)
	require.Equal(t, BasisPoints, report.HighCoverageBps)
	require.Equal(t, BasisPoints, report.MediumCoverageBps)
	require.Equal(t, BasisPoints, report.LowerCoverageBps)
	require.Contains(t, report.Summary, "backlog_high=10/10")
	require.Contains(t, report.Summary, "backlog_medium=7/7")
	require.Contains(t, report.Summary, "backlog_lower=6/6")

	for _, item := range report.Items {
		require.True(t, item.Tracked)
		require.NotEmpty(t, item.Description)
		require.NotEmpty(t, item.Evidence)
		if item.Priority == EconomicBacklogPriorityHigh && !item.LocalOnly {
			require.True(t, item.RequiresTests)
			if item.ID != EconomicBacklogFeeAllocationInvariantTests && item.ID != EconomicBacklogSlashingRouteInvariantTests {
				require.True(t, item.RequiresTelemetry || item.RequiresQuery)
			}
		}
	}
}

func TestDefaultEconomicBacklogTracksLocalEconomicsExclude(t *testing.T) {
	report := BuildEconomicEngineeringBacklogReport(nil)
	var exclude EconomicEngineeringBacklogItem
	for _, item := range report.Items {
		if item.ID == EconomicBacklogAddEconomicsLocalExclude {
			exclude = item
			break
		}
	}

	require.Equal(t, EconomicBacklogPriorityHigh, exclude.Priority)
	require.True(t, exclude.LocalOnly)
	require.Contains(t, exclude.Evidence, ".git/info/exclude:/ECONOMICS.md")
}

func TestEconomicBacklogRejectsMissingDuplicateAndWrongPriority(t *testing.T) {
	items := DefaultEconomicEngineeringBacklog()
	items = append(items[:1], items[2:]...)
	items = append(items, items[0])
	for i := range items {
		if items[i].ID == EconomicBacklogAdaptiveInflationController {
			items[i].Priority = EconomicBacklogPriorityHigh
		}
	}

	report := BuildEconomicEngineeringBacklogReport(items)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicBacklogEpochEconomicReportDataModel+":missing_required_backlog_item")
	require.Contains(t, report.Failed, EconomicBacklogAddEconomicsLocalExclude+":duplicate_backlog_item")
	require.Contains(t, report.Failed, EconomicBacklogAdaptiveInflationController+":wrong_backlog_priority")
	require.Less(t, report.HighCoverageBps, BasisPoints)
	require.Less(t, report.LowerCoverageBps, BasisPoints)
}

func TestEconomicBacklogRejectsUntrackedAndUnderspecifiedHighPriorityItems(t *testing.T) {
	items := DefaultEconomicEngineeringBacklog()
	for i := range items {
		switch items[i].ID {
		case EconomicBacklogEpochEconomicReportDataModel:
			items[i].Tracked = false
		case EconomicBacklogBurnFeeDistribution:
			items[i].RequiresTests = false
		case EconomicBacklogDeflationGuard:
			items[i].RequiresTelemetry = false
			items[i].RequiresQuery = false
		case EconomicBacklogStateGrowthTelemetry:
			items[i].Evidence = nil
		}
	}

	report := BuildEconomicEngineeringBacklogReport(items)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicBacklogEpochEconomicReportDataModel+":not_tracked")
	require.Contains(t, report.Failed, EconomicBacklogBurnFeeDistribution+":high_priority_requires_tests")
	require.Contains(t, report.Failed, EconomicBacklogDeflationGuard+":high_priority_requires_telemetry_or_query")
	require.Contains(t, report.Failed, EconomicBacklogStateGrowthTelemetry+":evidence_missing")
	require.Less(t, report.HighCoverageBps, BasisPoints)
}

func TestDefaultEconomicNonGoalsAreEnforced(t *testing.T) {
	report := BuildEconomicNonGoalReport(nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.NonGoals, 8)
	require.Equal(t, 8, report.Required)
	require.Equal(t, 8, report.Covered)
	require.Equal(t, BasisPoints, report.CoverageBps)
	require.Contains(t, report.Summary, "economic_non_goals=8/8")

	for _, nonGoal := range report.NonGoals {
		require.True(t, nonGoal.Tracked)
		require.NotEmpty(t, nonGoal.Statement)
		require.NotEmpty(t, nonGoal.Enforcement)
		if nonGoal.ID == EconomicNonGoalNondeterministicConsensusReputation {
			require.True(t, nonGoal.DeterminismGuard)
		}
		if nonGoal.ID == EconomicNonGoalBurnOverSecurityBudget {
			require.True(t, nonGoal.SecurityBudgetGuard)
		}
	}
}

func TestEconomicNonGoalsRejectMissingDuplicateAndMissingGuards(t *testing.T) {
	nonGoals := DefaultEconomicNonGoals()
	nonGoals = append(nonGoals[:1], nonGoals[2:]...)
	nonGoals = append(nonGoals, nonGoals[0])
	for i := range nonGoals {
		switch nonGoals[i].ID {
		case EconomicNonGoalNondeterministicConsensusReputation:
			nonGoals[i].DeterminismGuard = false
		case EconomicNonGoalBurnOverSecurityBudget:
			nonGoals[i].SecurityBudgetGuard = false
		case EconomicNonGoalUnboundedControllers:
			nonGoals[i].Enforcement = nil
		}
	}

	report := BuildEconomicNonGoalReport(nonGoals)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicNonGoalExternalValidatorRewardAssets+":missing_required_non_goal")
	require.Contains(t, report.Failed, EconomicNonGoalSecondStakingAsset+":duplicate_non_goal")
	require.Contains(t, report.Failed, EconomicNonGoalNondeterministicConsensusReputation+":determinism_guard_missing")
	require.Contains(t, report.Failed, EconomicNonGoalBurnOverSecurityBudget+":security_budget_guard_missing")
	require.Contains(t, report.Failed, EconomicNonGoalUnboundedControllers+":enforcement_missing")
	require.Less(t, report.CoverageBps, BasisPoints)
}
