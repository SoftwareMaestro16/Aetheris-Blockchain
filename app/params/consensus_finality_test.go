package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsensusFinalityReportAcceptsRequiredTargets(t *testing.T) {
	report := validConsensusFinalityReport()

	require.NoError(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report))
}

func TestConsensusFinalityReportRejectsUnstableHundredValidatorLocalnet(t *testing.T) {
	report := validConsensusFinalityReport()
	report.LocalnetStable = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "100-128 validator localnet")
}

func TestConsensusFinalityReportRejectsOneSecondBlocks(t *testing.T) {
	report := validConsensusFinalityReport()
	report.ObservedBlockTimeMinSeconds = 1
	report.ObservedBlockTimeMaxSeconds = 2

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "1-2 second block targets")
}

func TestConsensusFinalityReportRejectsFinalityOutsideBounds(t *testing.T) {
	report := validConsensusFinalityReport()
	report.NormalFinalitySeconds = 16
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "normal finality")

	report = validConsensusFinalityReport()
	report.StressFinalitySeconds = 91
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "stress finality")

	report = validConsensusFinalityReport()
	report.WorstFinalitySeconds = 121
	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "worst finality")
}

func TestConsensusFinalityReportRequiresDegradedLivenessWithHealthyTwoThirds(t *testing.T) {
	report := validConsensusFinalityReport()
	report.HealthyVotingPowerBps = AetraHealthyVotingPowerBps
	report.LivenessPreserved = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "2/3 voting power")
}

func TestConsensusFinalityReportRequiresTestnetReportMeasurements(t *testing.T) {
	report := validConsensusFinalityReport()
	report.IncludedInTestnetReport = false

	require.ErrorContains(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report), "testnet reports")
}

func TestConsensusFinalityReportValidatesMatureBlockTimeRange(t *testing.T) {
	report := validConsensusFinalityReport()
	report.ValidatorCount = 300
	report.ObservedBlockTimeMinSeconds = 7
	report.ObservedBlockTimeMaxSeconds = 8

	require.NoError(t, ValidateConsensusFinalityReport(DefaultNetworkProfile(), report))
}

func validConsensusFinalityReport() ConsensusFinalityReport {
	return ConsensusFinalityReport{
		ValidatorCount:              100,
		BlocksObserved:              100,
		LocalnetStable:              true,
		LoadProfileExecuted:         true,
		ObservedBlockTimeMinSeconds: 5,
		ObservedBlockTimeMaxSeconds: 6,
		NormalFinalitySeconds:       10,
		StressFinalitySeconds:       60,
		WorstFinalitySeconds:        90,
		DegradedScenarioExecuted:    true,
		HealthyVotingPowerBps:       AetraHealthyVotingPowerBps,
		LivenessPreserved:           true,
		IncludedInTestnetReport:     true,
	}
}
