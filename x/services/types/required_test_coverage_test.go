package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceRequiredTestCoverageCoversSection17(t *testing.T) {
	coverage, err := DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	require.NoError(t, coverage.Validate())
	require.NoError(t, ValidateServiceRequiredTestEvidence(coverage))
	require.NotEmpty(t, coverage.CoverageHash)

	requireServiceUnitCoverage(t, coverage, ServiceUnitServiceIDValidation)
	requireServiceUnitCoverage(t, coverage, ServiceUnitDescriptorHash)
	requireServiceUnitCoverage(t, coverage, ServiceUnitInterfaceHash)
	requireServiceUnitCoverage(t, coverage, ServiceUnitMethodIDValidation)
	requireServiceUnitCoverage(t, coverage, ServiceUnitCallIDDerivation)
	requireServiceUnitCoverage(t, coverage, ServiceUnitNonceReplay)
	requireServiceUnitCoverage(t, coverage, ServiceUnitIdempotencyKey)
	requireServiceUnitCoverage(t, coverage, ServiceUnitPaymentModel)
	requireServiceUnitCoverage(t, coverage, ServiceUnitTrustModel)
	requireServiceUnitCoverage(t, coverage, ServiceUnitReceiptHash)

	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationRegisterOnChainService)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationRegisterOffChainAnchor)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationRegisterMixedService)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationRegisterFogProvider)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationRegisterInterfaceBinding)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationResolveAETBinding)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationExecuteOnChainCall)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationAnchorOffChainResult)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationChallengeMixedResult)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationSettleEscrowPayment)
	requireServiceIntegrationCoverage(t, coverage, ServiceIntegrationGenerateReceiptProof)
}

func TestServiceRequiredTestCoverageRejectsMissingRequiredCases(t *testing.T) {
	coverage, err := DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.UnitTests = removeServiceUnitCoverageForTest(coverage.UnitTests, ServiceUnitNonceReplay)
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.ErrorContains(t, coverage.Validate(), "unit tests")

	coverage, err = DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.IntegrationTests = removeServiceIntegrationCoverageForTest(coverage.IntegrationTests, ServiceIntegrationChallengeMixedResult)
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.ErrorContains(t, coverage.Validate(), "integration tests")
}

func TestServiceRequiredTestCoverageRejectsHashTampering(t *testing.T) {
	testCase := newServiceUnitTestCase(ServiceUnitCallIDDerivation, "TestUnifiedCallIDDerivation", "ComputeUnifiedServiceCallID")
	testCase.TestName = "TestDifferentName"
	require.ErrorContains(t, testCase.Validate(), "hash mismatch")

	coverage, err := DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.UnitTests[0].EvidenceHash = ""
	coverage.UnitTests[0].CaseHash = ComputeServiceRequiredTestCaseHash(coverage.UnitTests[0])
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.Error(t, ValidateServiceRequiredTestEvidence(coverage))
}

func requireServiceUnitCoverage(t *testing.T, coverage ServiceRequiredTestCoverage, caseID ServiceUnitTestCaseID) {
	t.Helper()
	for _, testCase := range coverage.UnitTests {
		if testCase.UnitCase == caseID {
			require.Equal(t, ServiceRequiredTestUnit, testCase.Kind)
			require.NotEmpty(t, testCase.TestName)
			require.NotEmpty(t, testCase.EvidenceHash)
			return
		}
	}
	t.Fatalf("missing unit coverage %s", caseID)
}

func requireServiceIntegrationCoverage(t *testing.T, coverage ServiceRequiredTestCoverage, caseID ServiceIntegrationTestCaseID) {
	t.Helper()
	for _, testCase := range coverage.IntegrationTests {
		if testCase.Integration == caseID {
			require.Equal(t, ServiceRequiredTestIntegration, testCase.Kind)
			require.NotEmpty(t, testCase.TestName)
			require.NotEmpty(t, testCase.EvidenceHash)
			return
		}
	}
	t.Fatalf("missing integration coverage %s", caseID)
}

func removeServiceUnitCoverageForTest(testCases []ServiceRequiredTestCase, target ServiceUnitTestCaseID) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if testCase.UnitCase != target {
			out = append(out, testCase)
		}
	}
	return out
}

func removeServiceIntegrationCoverageForTest(testCases []ServiceRequiredTestCase, target ServiceIntegrationTestCaseID) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if testCase.Integration != target {
			out = append(out, testCase)
		}
	}
	return out
}
