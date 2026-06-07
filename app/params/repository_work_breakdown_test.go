package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraRepoProtoWorkCoversSection321(t *testing.T) {
	evidence := DefaultAetraRepoProtoWorkEvidence()

	report := BuildAetraRepoProtoWorkReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraRepoAreaProto, report.Area)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineMessages)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineQueryServices)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineTxServices)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineGenesis)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskDefineParams)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskRunCodeGeneration)
	require.Contains(t, evidence.Tasks, AetraRepoProtoTaskBreakingChangeChecks)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestGeneratedCodeCompiles)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestLintPasses)
	require.Contains(t, evidence.Tests, AetraRepoProtoTestServiceRegistration)
	require.NoError(t, ValidateAetraRepoProtoWork(evidence))
}

func TestAetraRepoProtoWorkRejectsMissingTasksAndTests(t *testing.T) {
	evidence := DefaultAetraRepoProtoWorkEvidence()
	evidence.Area = "x/proto"
	evidence.Tasks = removeRepoWorkItem(evidence.Tasks,
		AetraRepoProtoTaskDefineMessages,
		AetraRepoProtoTaskDefineQueryServices,
		AetraRepoProtoTaskRunCodeGeneration,
	)
	evidence.Tests = removeRepoWorkItem(evidence.Tests,
		AetraRepoProtoTestGeneratedCodeCompiles,
		AetraRepoProtoTestServiceRegistration,
	)
	evidence.Tasks = append(evidence.Tasks, AetraRepoProtoTaskDefineParams, "manual_proto_note")
	evidence.Tests = append(evidence.Tests, AetraRepoProtoTestLintPasses, "manual_buf_review_only")

	report := BuildAetraRepoProtoWorkReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "area_must_be_"+AetraRepoAreaProto)
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineMessages+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineQueryServices+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskRunCodeGeneration+":missing")
	require.Contains(t, report.Failed, "tasks."+AetraRepoProtoTaskDefineParams+":duplicate")
	require.Contains(t, report.Failed, "tasks.manual_proto_note:unexpected")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestGeneratedCodeCompiles+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestServiceRegistration+":missing")
	require.Contains(t, report.Failed, "tests."+AetraRepoProtoTestLintPasses+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_buf_review_only:unexpected")
	require.Error(t, ValidateAetraRepoProtoWork(evidence))
}

func removeRepoWorkItem(items []string, targets ...string) []string {
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
