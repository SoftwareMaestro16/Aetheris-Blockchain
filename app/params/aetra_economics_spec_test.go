package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraEconomicsSpecCoversModulePurpose(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()

	report := BuildAetraEconomicsSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 5, report.Required)
	require.NoError(t, ValidateAetraEconomicsSpec(evidence))
}

func TestAetraEconomicsSpecRejectsMissingPurposeComponents(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()
	evidence.LowModerateInflation = false
	evidence.FeeBurn = false
	evidence.TreasuryAllocation = false
	evidence.RewardSmoothing = false
	evidence.TransparentAPRModel = false

	report := BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraEconomicsPurposeLowModerateInflation)
	require.Contains(t, report.Failed, AetraEconomicsPurposeFeeBurn)
	require.Contains(t, report.Failed, AetraEconomicsPurposeTreasuryAllocation)
	require.Contains(t, report.Failed, AetraEconomicsPurposeRewardSmoothing)
	require.Contains(t, report.Failed, AetraEconomicsPurposeTransparentAPRModel)
	require.Error(t, ValidateAetraEconomicsSpec(evidence))
}

func TestAetraEconomicsSpecRejectsWrongModuleIdentity(t *testing.T) {
	evidence := DefaultAetraEconomicsSpecEvidence()
	evidence.ModuleName = ""

	report := BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")

	evidence.ModuleName = "x/economics"
	report = BuildAetraEconomicsSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
}

func TestDefaultAetraEconomicsResponsibilitiesCoverSection231(t *testing.T) {
	evidence := DefaultAetraEconomicsResponsibilitiesEvidence()

	report := BuildAetraEconomicsResponsibilitiesReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.NoError(t, ValidateAetraEconomicsResponsibilities(evidence))
}

func TestAetraEconomicsResponsibilitiesRejectMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraEconomicsResponsibilitiesEvidence()
	evidence.ModuleName = "x/economics"
	evidence.CalculatesDynamicInflation = false
	evidence.TracksBondedRatio = false
	evidence.EstimatesStakingAPR = false
	evidence.SplitsFees = false
	evidence.BurnsConfiguredFeeShare = false
	evidence.SendsConfiguredShareToDistributionRewards = false
	evidence.SendsConfiguredShareToTreasury = false
	evidence.SmoothsRewardChanges = false
	evidence.ExposesEconomicMetrics = false
	evidence.ProtectsSupplyInvariants = false

	report := BuildAetraEconomicsResponsibilitiesReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityDynamicInflation)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityBondedRatio)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityStakingAPR)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityFeeSplit)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityBurnFeeShare)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityRewardsShare)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityTreasuryShare)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityRewardSmoothing)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilityEconomicMetrics)
	require.Contains(t, report.Failed, AetraEconomicsResponsibilitySupplyInvariants)
	require.Error(t, ValidateAetraEconomicsResponsibilities(evidence))
}

func TestDefaultAetraEconomicsStateSpecCoversSection232(t *testing.T) {
	evidence := DefaultAetraEconomicsStateSpecEvidence()

	report := BuildAetraEconomicsStateSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 24, report.Required)
	require.NoError(t, ValidateAetraEconomicsStateSpec(evidence))
}

func TestAetraEconomicsStateSpecRejectsMissingFields(t *testing.T) {
	evidence := DefaultAetraEconomicsStateSpecEvidence()
	evidence.ParamsFields = removeEconomicsString(evidence.ParamsFields, AetraEconomicsStateParamInflationChangeRateBps)
	evidence.EpochEconomicsFields = removeEconomicsString(evidence.EpochEconomicsFields, AetraEconomicsStateEpochFeesToTreasury)
	evidence.SupplyStatsFields = removeEconomicsString(evidence.SupplyStatsFields, AetraEconomicsStateSupplyNetIssuance)

	report := BuildAetraEconomicsStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraEconomicsStateParams+"."+AetraEconomicsStateParamInflationChangeRateBps+":missing")
	require.Contains(t, report.Failed, AetraEconomicsStateEpochEconomics+"."+AetraEconomicsStateEpochFeesToTreasury+":missing")
	require.Contains(t, report.Failed, AetraEconomicsStateSupplyStats+"."+AetraEconomicsStateSupplyNetIssuance+":missing")
	require.Error(t, ValidateAetraEconomicsStateSpec(evidence))
}

func TestAetraEconomicsStateSpecRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraEconomicsStateSpecEvidence()
	evidence.ModuleName = "x/economics"
	evidence.ParamsFields = append(evidence.ParamsFields, AetraEconomicsStateParamBurnFeeShareBps, "FloatingPointInflation")
	evidence.EpochEconomicsFields = append(evidence.EpochEconomicsFields, AetraEconomicsStateEpochFeesBurned, "SubjectiveRewardScore")
	evidence.SupplyStatsFields = append(evidence.SupplyStatsFields, AetraEconomicsStateSupplyTotalBurned, "ExternalSupplyOracle")

	report := BuildAetraEconomicsStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
	require.Contains(t, report.Failed, AetraEconomicsStateParams+"."+AetraEconomicsStateParamBurnFeeShareBps+":duplicate")
	require.Contains(t, report.Failed, AetraEconomicsStateParams+".FloatingPointInflation:unexpected")
	require.Contains(t, report.Failed, AetraEconomicsStateEpochEconomics+"."+AetraEconomicsStateEpochFeesBurned+":duplicate")
	require.Contains(t, report.Failed, AetraEconomicsStateEpochEconomics+".SubjectiveRewardScore:unexpected")
	require.Contains(t, report.Failed, AetraEconomicsStateSupplyStats+"."+AetraEconomicsStateSupplyTotalBurned+":duplicate")
	require.Contains(t, report.Failed, AetraEconomicsStateSupplyStats+".ExternalSupplyOracle:unexpected")
	require.Error(t, ValidateAetraEconomicsStateSpec(evidence))
}

func TestDefaultAetraEconomicsInflationCurveCoversSection233(t *testing.T) {
	evidence := DefaultAetraEconomicsInflationCurveEvidence()

	report := BuildAetraEconomicsInflationCurveReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 8, report.Required)
	require.NoError(t, ValidateAetraEconomicsInflationCurve(evidence))
}

func TestAetraEconomicsInflationCurveRejectsMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraEconomicsInflationCurveEvidence()
	evidence.ModuleName = ""
	evidence.BondedRatioBelowTargetIncreasesInflation = false
	evidence.BondedRatioAboveTargetDecreasesInflation = false
	evidence.InflationNeverBelowMin = false
	evidence.InflationNeverAboveMax = false
	evidence.InflationChangePerEpochBounded = false
	evidence.NoFloatingPoint = false
	evidence.NoPerBlockInstability = false
	evidence.AllCalculationsDeterministic = false

	report := BuildAetraEconomicsInflationCurveReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveBelowTargetIncreases)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveAboveTargetDecreases)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveNeverBelowMin)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveNeverAboveMax)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveEpochChangeBounded)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveNoFloatingPoint)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveNoPerBlockInstability)
	require.Contains(t, report.Failed, AetraEconomicsInflationCurveDeterministic)
	require.Error(t, ValidateAetraEconomicsInflationCurve(evidence))
}

func TestDefaultAetraEconomicsFeeSplitRulesCoverSection234(t *testing.T) {
	evidence := DefaultAetraEconomicsFeeSplitRulesEvidence()

	report := BuildAetraEconomicsFeeSplitRulesReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 9, report.Required)
	require.NoError(t, ValidateAetraEconomicsFeeSplitRules(evidence))
}

func TestAetraEconomicsFeeSplitRulesRejectMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraEconomicsFeeSplitRulesEvidence()
	evidence.ModuleName = "x/economics"
	evidence.FeeSplitSumsToBasisPoints = false
	evidence.RecommendedBurnRange = false
	evidence.RecommendedRewardRange = false
	evidence.RecommendedTreasuryRange = false
	evidence.RejectsInvalidSum = false
	evidence.RejectsNegativeShares = false
	evidence.RejectsBurnAboveGovernanceMax = false
	evidence.RejectsTreasuryAboveGovernanceMax = false
	evidence.RejectsZeroRewardsWithoutEmergency = false

	report := BuildAetraEconomicsFeeSplitRulesReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitSumToBasisPoints)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRecommendedBurnRange)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRecommendedRewardRange)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRecommendedTreasuryRange)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRejectsInvalidSum)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRejectsNegativeShares)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRejectsBurnAboveGovernanceMax)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRejectsTreasuryAboveMax)
	require.Contains(t, report.Failed, AetraEconomicsFeeSplitRejectsZeroRewards)
	require.Error(t, ValidateAetraEconomicsFeeSplitRules(evidence))
}

func TestDefaultAetraEconomicsAPRQueryCoversSection235(t *testing.T) {
	evidence := DefaultAetraEconomicsAPRQueryEvidence()

	report := BuildAetraEconomicsAPRQueryReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 7, report.Required)
	require.NoError(t, ValidateAetraEconomicsAPRQuery(evidence))
}

func TestAetraEconomicsAPRQueryRejectsMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraEconomicsAPRQueryEvidence()
	evidence.ModuleName = ""
	evidence.InflationOnlyAPR = false
	evidence.FeeAdjustedAPR = false
	evidence.ValidatorCommissionImpact = false
	evidence.EstimatedDelegatorAPR = false
	evidence.EstimatedValidatorGrossAPR = false
	evidence.EstimatedValidatorNetAPR = false
	evidence.LabeledAsEstimate = false

	report := BuildAetraEconomicsAPRQueryReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryInflationOnly)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryFeeAdjusted)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryCommissionImpact)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryDelegatorEstimate)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryValidatorGross)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryValidatorNet)
	require.Contains(t, report.Failed, AetraEconomicsAPRQueryLabeledAsEstimate)
	require.Error(t, ValidateAetraEconomicsAPRQuery(evidence))
}

func TestDefaultAetraEconomicsTestingRequirementsCoverSection236(t *testing.T) {
	evidence := DefaultAetraEconomicsTestingRequirementsEvidence()

	report := BuildAetraEconomicsTestingRequirementsReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraEconomicsModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 14, report.Required)
	require.NoError(t, ValidateAetraEconomicsTestingRequirements(evidence))
}

func TestAetraEconomicsTestingRequirementsRejectMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraEconomicsTestingRequirementsEvidence()
	evidence.ModuleName = "x/economics"
	evidence.InflationIncreasesBelowTarget = false
	evidence.InflationDecreasesAboveTarget = false
	evidence.InflationWithinMinMax = false
	evidence.InflationChangeRateBounded = false
	evidence.FeeSplitExactAccounting = false
	evidence.BurnAccounting = false
	evidence.TreasuryAccounting = false
	evidence.RewardsAccounting = false
	evidence.APRMath = false
	evidence.ZeroFeeBlockHandling = false
	evidence.HighFeeBlockHandling = false
	evidence.ExportImportEconomicsState = false
	evidence.SupplyInvariantAfterManyEpochs = false
	evidence.GovernanceInvalidParamsRejected = false

	report := BuildAetraEconomicsTestingRequirementsReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraEconomicsModuleName)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestInflationBelowTarget)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestInflationAboveTarget)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestInflationMinMax)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestInflationChangeBounded)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestFeeSplitAccounting)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestBurnAccounting)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestTreasuryAccounting)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestRewardsAccounting)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestAPRMath)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestZeroFeeBlock)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestHighFeeBlock)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestExportImportState)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestSupplyInvariantManyEpochs)
	require.Contains(t, report.Failed, AetraEconomicsRequiredTestGovernanceInvalidParams)
	require.Error(t, ValidateAetraEconomicsTestingRequirements(evidence))
}

func removeEconomicsString(values []string, targets ...string) []string {
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
