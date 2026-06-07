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
