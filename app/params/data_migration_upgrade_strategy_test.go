package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraUpgradeStrategyCoversSection31(t *testing.T) {
	evidence := DefaultAetraUpgradeStrategyEvidence("x/aetra-economics")

	report := BuildAetraUpgradeStrategyReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, "x/aetra-economics", report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 14, report.Required)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementStoreKeyDecision)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementGenesisImportExport)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementMigrationHandler)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementVersionMapUpdate)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementUpgradeTest)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementRollbackNotes)
	require.Contains(t, evidence.Requirements, AetraUpgradeRequirementOperatorInstructions)
	require.Contains(t, evidence.Tests, AetraMigrationTestOldGenesisImports)
	require.Contains(t, evidence.Tests, AetraMigrationTestInitializesParams)
	require.Contains(t, evidence.Tests, AetraMigrationTestPreservesBalances)
	require.Contains(t, evidence.Tests, AetraMigrationTestPreservesStakingState)
	require.Contains(t, evidence.Tests, AetraMigrationTestPreservesSlashingState)
	require.Contains(t, evidence.Tests, AetraMigrationTestPreservesContractState)
	require.Contains(t, evidence.Tests, AetraMigrationTestDeterministicAppHash)
	require.NoError(t, ValidateAetraUpgradeStrategy(evidence))
}

func TestAetraUpgradeStrategyRejectsMissingRequirementsAndTests(t *testing.T) {
	evidence := DefaultAetraUpgradeStrategyEvidence("")
	evidence.Requirements = removeUpgradeItem(evidence.Requirements,
		AetraUpgradeRequirementStoreKeyDecision,
		AetraUpgradeRequirementMigrationHandler,
		AetraUpgradeRequirementOperatorInstructions,
	)
	evidence.Tests = removeUpgradeItem(evidence.Tests,
		AetraMigrationTestOldGenesisImports,
		AetraMigrationTestPreservesBalances,
		AetraMigrationTestDeterministicAppHash,
	)
	evidence.Requirements = append(evidence.Requirements, AetraUpgradeRequirementVersionMapUpdate, "manual_wiki_note")
	evidence.Tests = append(evidence.Tests, AetraMigrationTestInitializesParams, "manual_upgrade_smoke_only")

	report := BuildAetraUpgradeStrategyReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "requirements."+AetraUpgradeRequirementStoreKeyDecision+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraUpgradeRequirementMigrationHandler+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraUpgradeRequirementOperatorInstructions+":missing")
	require.Contains(t, report.Failed, "requirements."+AetraUpgradeRequirementVersionMapUpdate+":duplicate")
	require.Contains(t, report.Failed, "requirements.manual_wiki_note:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraMigrationTestOldGenesisImports+":missing")
	require.Contains(t, report.Failed, "tests."+AetraMigrationTestPreservesBalances+":missing")
	require.Contains(t, report.Failed, "tests."+AetraMigrationTestDeterministicAppHash+":missing")
	require.Contains(t, report.Failed, "tests."+AetraMigrationTestInitializesParams+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_upgrade_smoke_only:unexpected")
	require.Error(t, ValidateAetraUpgradeStrategy(evidence))
}

func removeUpgradeItem(items []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !targetSet[item] {
			out = append(out, item)
		}
	}
	return out
}
