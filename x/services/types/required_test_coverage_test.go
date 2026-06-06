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

	requireServiceInvariantCoverage(t, coverage, ServiceInvariantDescriptorStoredHash)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantInterfaceRegisteredHash)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantActiveServiceInterface)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantCallReceiptServiceMethod)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantPaymentSettlementEscrowLimit)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantProviderCollateralNonNegative)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantExpiredServiceRejectsCalls)
	requireServiceInvariantCoverage(t, coverage, ServiceInvariantReceiptRootCommittedReceipts)

	requireServiceFuzzCoverage(t, coverage, ServiceFuzzMalformedDescriptors)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzMalformedInterfaceSchemas)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzLargePayloadCalls)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzDuplicateNonces)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzDuplicateIdempotencyKeys)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzForgedProviderSignatures)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzInvalidResultAnchors)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzInvalidDisputeProofs)
	requireServiceFuzzCoverage(t, coverage, ServiceFuzzPaymentEdgeCases)

	requireServicePerformanceCoverage(t, coverage, ServicePerformanceRegistryLookupLatency)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceInterfaceLookupLatency)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceServiceCallEnqueue)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceOnChainExecutionThroughput)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceReceiptAnchoringThroughput)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceReceiptProofLatency)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceProviderLookupLatency)
	requireServicePerformanceCoverage(t, coverage, ServicePerformanceBlockSTMConflictRate)
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

	coverage, err = DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.InvariantTests = removeServiceInvariantCoverageForTest(coverage.InvariantTests, ServiceInvariantReceiptRootCommittedReceipts)
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.ErrorContains(t, coverage.Validate(), "invariant tests")

	coverage, err = DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.FuzzTests = removeServiceFuzzCoverageForTest(coverage.FuzzTests, ServiceFuzzInvalidDisputeProofs)
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.ErrorContains(t, coverage.Validate(), "fuzz tests")

	coverage, err = DefaultServiceRequiredTestCoverage()
	require.NoError(t, err)
	coverage.PerformanceTests = removeServicePerformanceCoverageForTest(coverage.PerformanceTests, ServicePerformanceBlockSTMConflictRate)
	coverage.CoverageHash = ComputeServiceRequiredTestCoverageHash(coverage)
	require.ErrorContains(t, coverage.Validate(), "performance tests")
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

func requireServiceInvariantCoverage(t *testing.T, coverage ServiceRequiredTestCoverage, caseID ServiceInvariantTestCaseID) {
	t.Helper()
	for _, testCase := range coverage.InvariantTests {
		if testCase.Invariant == caseID {
			require.Equal(t, ServiceRequiredTestInvariant, testCase.Kind)
			require.NotEmpty(t, testCase.TestName)
			require.NotEmpty(t, testCase.EvidenceHash)
			return
		}
	}
	t.Fatalf("missing invariant coverage %s", caseID)
}

func requireServiceFuzzCoverage(t *testing.T, coverage ServiceRequiredTestCoverage, caseID ServiceFuzzTestCaseID) {
	t.Helper()
	for _, testCase := range coverage.FuzzTests {
		if testCase.Fuzz == caseID {
			require.Equal(t, ServiceRequiredTestFuzz, testCase.Kind)
			require.NotEmpty(t, testCase.TestName)
			require.NotEmpty(t, testCase.EvidenceHash)
			return
		}
	}
	t.Fatalf("missing fuzz coverage %s", caseID)
}

func requireServicePerformanceCoverage(t *testing.T, coverage ServiceRequiredTestCoverage, caseID ServicePerformanceTestCaseID) {
	t.Helper()
	for _, testCase := range coverage.PerformanceTests {
		if testCase.Performance == caseID {
			require.Equal(t, ServiceRequiredTestPerformance, testCase.Kind)
			require.NotEmpty(t, testCase.TestName)
			require.NotEmpty(t, testCase.EvidenceHash)
			return
		}
	}
	t.Fatalf("missing performance coverage %s", caseID)
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

func removeServiceInvariantCoverageForTest(testCases []ServiceRequiredTestCase, target ServiceInvariantTestCaseID) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if testCase.Invariant != target {
			out = append(out, testCase)
		}
	}
	return out
}

func removeServiceFuzzCoverageForTest(testCases []ServiceRequiredTestCase, target ServiceFuzzTestCaseID) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if testCase.Fuzz != target {
			out = append(out, testCase)
		}
	}
	return out
}

func removeServicePerformanceCoverageForTest(testCases []ServiceRequiredTestCase, target ServicePerformanceTestCaseID) []ServiceRequiredTestCase {
	out := make([]ServiceRequiredTestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if testCase.Performance != target {
			out = append(out, testCase)
		}
	}
	return out
}
