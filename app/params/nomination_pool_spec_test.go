package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraNominationPoolModelCoversSection261(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()

	report := BuildAetraNominationPoolModelReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraNominationPoolModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 18, report.Required)
	require.NoError(t, ValidateAetraNominationPoolModel(evidence))
}

func TestAetraNominationPoolModelRejectsMissingRequiredFieldsAndRisks(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()
	evidence.ModuleName = ""
	evidence.PoolFields = removeNominationPoolString(evidence.PoolFields, AetraNominationPoolFieldCreatedHeight)
	evidence.PoolDelegationFields = removeNominationPoolString(evidence.PoolDelegationFields, AetraNominationPoolFieldPrincipalEstimate)
	evidence.AccessibilityRiskAcknowledged = false
	evidence.AccountingRiskAcknowledged = false
	evidence.CentralizationRiskAcknowledged = false
	evidence.CurrentImplementationMapped = false

	report := BuildAetraNominationPoolModelReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+"."+AetraNominationPoolFieldCreatedHeight+":missing")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+"."+AetraNominationPoolFieldPrincipalEstimate+":missing")
	require.Contains(t, report.Failed, AetraNominationPoolRiskAccessibility)
	require.Contains(t, report.Failed, AetraNominationPoolRiskAccounting)
	require.Contains(t, report.Failed, AetraNominationPoolRiskCentralization)
	require.Contains(t, report.Failed, AetraNominationPoolImplementationMap)
	require.Error(t, ValidateAetraNominationPoolModel(evidence))
}

func TestAetraNominationPoolModelRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()
	evidence.ModuleName = "x/nomination-pool"
	evidence.PoolFields = append(evidence.PoolFields, AetraNominationPoolFieldPoolID, "OperatorKycStatus")
	evidence.PoolDelegationFields = append(evidence.PoolDelegationFields, AetraNominationPoolFieldShares, "LocalUiEstimate")

	report := BuildAetraNominationPoolModelReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+"."+AetraNominationPoolFieldPoolID+":duplicate")
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+".OperatorKycStatus:unexpected")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+"."+AetraNominationPoolFieldShares+":duplicate")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+".LocalUiEstimate:unexpected")
	require.Error(t, ValidateAetraNominationPoolModel(evidence))
}

func removeNominationPoolString(values []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}

	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if !targetSet[value] {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
