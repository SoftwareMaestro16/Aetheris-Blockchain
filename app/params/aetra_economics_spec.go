package params

import (
	"fmt"
	"sort"
)

const (
	AetraEconomicsModuleName = "x/aetra-economics"

	AetraEconomicsPurposeLowModerateInflation = "low_moderate_inflation"
	AetraEconomicsPurposeFeeBurn              = "fee_burn"
	AetraEconomicsPurposeTreasuryAllocation   = "treasury_allocation"
	AetraEconomicsPurposeRewardSmoothing      = "reward_smoothing"
	AetraEconomicsPurposeTransparentAPRModel  = "transparent_apr_model"

	AetraEconomicsResponsibilityDynamicInflation = "calculate_dynamic_inflation"
	AetraEconomicsResponsibilityBondedRatio      = "track_bonded_ratio"
	AetraEconomicsResponsibilityStakingAPR       = "estimate_staking_apr"
	AetraEconomicsResponsibilityFeeSplit         = "split_fees"
	AetraEconomicsResponsibilityBurnFeeShare     = "burn_configured_fee_share"
	AetraEconomicsResponsibilityRewardsShare     = "send_configured_share_to_distribution_rewards"
	AetraEconomicsResponsibilityTreasuryShare    = "send_configured_share_to_treasury"
	AetraEconomicsResponsibilityRewardSmoothing  = "smooth_reward_changes"
	AetraEconomicsResponsibilityEconomicMetrics  = "expose_economic_metrics"
	AetraEconomicsResponsibilitySupplyInvariants = "protect_supply_invariants"

	AetraEconomicsStateParams         = "Params"
	AetraEconomicsStateEpochEconomics = "EpochEconomics"
	AetraEconomicsStateSupplyStats    = "SupplyStats"

	AetraEconomicsStateParamInflationMinBps        = "InflationMinBps"
	AetraEconomicsStateParamInflationMaxBps        = "InflationMaxBps"
	AetraEconomicsStateParamInflationChangeRateBps = "InflationChangeRateBps"
	AetraEconomicsStateParamTargetBondedRatioBps   = "TargetBondedRatioBps"
	AetraEconomicsStateParamBurnFeeShareBps        = "BurnFeeShareBps"
	AetraEconomicsStateParamRewardFeeShareBps      = "RewardFeeShareBps"
	AetraEconomicsStateParamTreasuryFeeShareBps    = "TreasuryFeeShareBps"
	AetraEconomicsStateParamRewardSmoothingEpochs  = "RewardSmoothingEpochs"
	AetraEconomicsStateParamAprTargetMinBps        = "AprTargetMinBps"
	AetraEconomicsStateParamAprTargetMaxBps        = "AprTargetMaxBps"

	AetraEconomicsStateEpochNumber          = "EpochNumber"
	AetraEconomicsStateEpochStartHeight     = "StartHeight"
	AetraEconomicsStateEpochEndHeight       = "EndHeight"
	AetraEconomicsStateEpochBondedRatioBps  = "BondedRatioBps"
	AetraEconomicsStateEpochInflationBps    = "InflationBps"
	AetraEconomicsStateEpochEstimatedAprBps = "EstimatedAprBps"
	AetraEconomicsStateEpochFeesCollected   = "FeesCollected"
	AetraEconomicsStateEpochFeesBurned      = "FeesBurned"
	AetraEconomicsStateEpochFeesToRewards   = "FeesToRewards"
	AetraEconomicsStateEpochFeesToTreasury  = "FeesToTreasury"
	AetraEconomicsStateEpochMintedRewards   = "MintedRewards"

	AetraEconomicsStateSupplyTotalMinted = "TotalMinted"
	AetraEconomicsStateSupplyTotalBurned = "TotalBurned"
	AetraEconomicsStateSupplyNetIssuance = "NetIssuance"
)

type AetraEconomicsSpecEvidence struct {
	ModuleName string

	LowModerateInflation bool
	FeeBurn              bool
	TreasuryAllocation   bool
	RewardSmoothing      bool
	TransparentAPRModel  bool
}

type AetraEconomicsSpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraEconomicsResponsibilitiesEvidence struct {
	ModuleName string

	CalculatesDynamicInflation                bool
	TracksBondedRatio                         bool
	EstimatesStakingAPR                       bool
	SplitsFees                                bool
	BurnsConfiguredFeeShare                   bool
	SendsConfiguredShareToDistributionRewards bool
	SendsConfiguredShareToTreasury            bool
	SmoothsRewardChanges                      bool
	ExposesEconomicMetrics                    bool
	ProtectsSupplyInvariants                  bool
}

type AetraEconomicsResponsibilitiesReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

type AetraEconomicsStateSpecEvidence struct {
	ModuleName string

	ParamsFields         []string
	EpochEconomicsFields []string
	SupplyStatsFields    []string
}

type AetraEconomicsStateSpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraEconomicsSpecEvidence() AetraEconomicsSpecEvidence {
	return AetraEconomicsSpecEvidence{
		ModuleName: AetraEconomicsModuleName,

		LowModerateInflation: true,
		FeeBurn:              true,
		TreasuryAllocation:   true,
		RewardSmoothing:      true,
		TransparentAPRModel:  true,
	}
}

func ValidateAetraEconomicsSpec(evidence AetraEconomicsSpecEvidence) error {
	report := BuildAetraEconomicsSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsSpecReport(evidence AetraEconomicsSpecEvidence) AetraEconomicsSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsPurposeLowModerateInflation, evidence.LowModerateInflation},
		{AetraEconomicsPurposeFeeBurn, evidence.FeeBurn},
		{AetraEconomicsPurposeTreasuryAllocation, evidence.TreasuryAllocation},
		{AetraEconomicsPurposeRewardSmoothing, evidence.RewardSmoothing},
		{AetraEconomicsPurposeTransparentAPRModel, evidence.TransparentAPRModel},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraEconomicsSpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func DefaultAetraEconomicsResponsibilitiesEvidence() AetraEconomicsResponsibilitiesEvidence {
	return AetraEconomicsResponsibilitiesEvidence{
		ModuleName: AetraEconomicsModuleName,

		CalculatesDynamicInflation:                true,
		TracksBondedRatio:                         true,
		EstimatesStakingAPR:                       true,
		SplitsFees:                                true,
		BurnsConfiguredFeeShare:                   true,
		SendsConfiguredShareToDistributionRewards: true,
		SendsConfiguredShareToTreasury:            true,
		SmoothsRewardChanges:                      true,
		ExposesEconomicMetrics:                    true,
		ProtectsSupplyInvariants:                  true,
	}
}

func ValidateAetraEconomicsResponsibilities(evidence AetraEconomicsResponsibilitiesEvidence) error {
	report := BuildAetraEconomicsResponsibilitiesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics responsibilities failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsResponsibilitiesReport(evidence AetraEconomicsResponsibilitiesEvidence) AetraEconomicsResponsibilitiesReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsResponsibilityDynamicInflation, evidence.CalculatesDynamicInflation},
		{AetraEconomicsResponsibilityBondedRatio, evidence.TracksBondedRatio},
		{AetraEconomicsResponsibilityStakingAPR, evidence.EstimatesStakingAPR},
		{AetraEconomicsResponsibilityFeeSplit, evidence.SplitsFees},
		{AetraEconomicsResponsibilityBurnFeeShare, evidence.BurnsConfiguredFeeShare},
		{AetraEconomicsResponsibilityRewardsShare, evidence.SendsConfiguredShareToDistributionRewards},
		{AetraEconomicsResponsibilityTreasuryShare, evidence.SendsConfiguredShareToTreasury},
		{AetraEconomicsResponsibilityRewardSmoothing, evidence.SmoothsRewardChanges},
		{AetraEconomicsResponsibilityEconomicMetrics, evidence.ExposesEconomicMetrics},
		{AetraEconomicsResponsibilitySupplyInvariants, evidence.ProtectsSupplyInvariants},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraEconomicsResponsibilitiesReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func DefaultAetraEconomicsStateSpecEvidence() AetraEconomicsStateSpecEvidence {
	return AetraEconomicsStateSpecEvidence{
		ModuleName: AetraEconomicsModuleName,
		ParamsFields: []string{
			AetraEconomicsStateParamInflationMinBps,
			AetraEconomicsStateParamInflationMaxBps,
			AetraEconomicsStateParamInflationChangeRateBps,
			AetraEconomicsStateParamTargetBondedRatioBps,
			AetraEconomicsStateParamBurnFeeShareBps,
			AetraEconomicsStateParamRewardFeeShareBps,
			AetraEconomicsStateParamTreasuryFeeShareBps,
			AetraEconomicsStateParamRewardSmoothingEpochs,
			AetraEconomicsStateParamAprTargetMinBps,
			AetraEconomicsStateParamAprTargetMaxBps,
		},
		EpochEconomicsFields: []string{
			AetraEconomicsStateEpochNumber,
			AetraEconomicsStateEpochStartHeight,
			AetraEconomicsStateEpochEndHeight,
			AetraEconomicsStateEpochBondedRatioBps,
			AetraEconomicsStateEpochInflationBps,
			AetraEconomicsStateEpochEstimatedAprBps,
			AetraEconomicsStateEpochFeesCollected,
			AetraEconomicsStateEpochFeesBurned,
			AetraEconomicsStateEpochFeesToRewards,
			AetraEconomicsStateEpochFeesToTreasury,
			AetraEconomicsStateEpochMintedRewards,
		},
		SupplyStatsFields: []string{
			AetraEconomicsStateSupplyTotalMinted,
			AetraEconomicsStateSupplyTotalBurned,
			AetraEconomicsStateSupplyNetIssuance,
		},
	}
}

func ValidateAetraEconomicsStateSpec(evidence AetraEconomicsStateSpecEvidence) error {
	report := BuildAetraEconomicsStateSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics state spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsStateSpecReport(evidence AetraEconomicsStateSpecEvidence) AetraEconomicsStateSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	requiredParams := requiredAetraEconomicsParamsFields()
	requiredEpoch := requiredAetraEconomicsEpochEconomicsFields()
	requiredSupply := requiredAetraEconomicsSupplyStatsFields()

	passedParams, failedParams := validateAetraEconomicsCatalog(AetraEconomicsStateParams, evidence.ParamsFields, requiredParams)
	passedEpoch, failedEpoch := validateAetraEconomicsCatalog(AetraEconomicsStateEpochEconomics, evidence.EpochEconomicsFields, requiredEpoch)
	passedSupply, failedSupply := validateAetraEconomicsCatalog(AetraEconomicsStateSupplyStats, evidence.SupplyStatsFields, requiredSupply)

	failed = append(failed, failedParams...)
	failed = append(failed, failedEpoch...)
	failed = append(failed, failedSupply...)
	sort.Strings(failed)
	return AetraEconomicsStateSpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(requiredParams) + len(requiredEpoch) + len(requiredSupply),
		Passed:     passedParams + passedEpoch + passedSupply,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func requiredAetraEconomicsParamsFields() []string {
	return []string{
		AetraEconomicsStateParamInflationMinBps,
		AetraEconomicsStateParamInflationMaxBps,
		AetraEconomicsStateParamInflationChangeRateBps,
		AetraEconomicsStateParamTargetBondedRatioBps,
		AetraEconomicsStateParamBurnFeeShareBps,
		AetraEconomicsStateParamRewardFeeShareBps,
		AetraEconomicsStateParamTreasuryFeeShareBps,
		AetraEconomicsStateParamRewardSmoothingEpochs,
		AetraEconomicsStateParamAprTargetMinBps,
		AetraEconomicsStateParamAprTargetMaxBps,
	}
}

func requiredAetraEconomicsEpochEconomicsFields() []string {
	return []string{
		AetraEconomicsStateEpochNumber,
		AetraEconomicsStateEpochStartHeight,
		AetraEconomicsStateEpochEndHeight,
		AetraEconomicsStateEpochBondedRatioBps,
		AetraEconomicsStateEpochInflationBps,
		AetraEconomicsStateEpochEstimatedAprBps,
		AetraEconomicsStateEpochFeesCollected,
		AetraEconomicsStateEpochFeesBurned,
		AetraEconomicsStateEpochFeesToRewards,
		AetraEconomicsStateEpochFeesToTreasury,
		AetraEconomicsStateEpochMintedRewards,
	}
}

func requiredAetraEconomicsSupplyStatsFields() []string {
	return []string{
		AetraEconomicsStateSupplyTotalMinted,
		AetraEconomicsStateSupplyTotalBurned,
		AetraEconomicsStateSupplyNetIssuance,
	}
}

func validateAetraEconomicsCatalog(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, item := range required {
		requiredSet[item] = true
	}
	seen := map[string]bool{}
	for _, item := range actual {
		if item == "" {
			failed = append(failed, group+".item_required")
			continue
		}
		if seen[item] {
			failed = append(failed, group+"."+item+":duplicate")
			continue
		}
		seen[item] = true
		if !requiredSet[item] {
			failed = append(failed, group+"."+item+":unexpected")
		}
	}
	passed := 0
	for _, item := range required {
		if seen[item] {
			passed++
		} else {
			failed = append(failed, group+"."+item+":missing")
		}
	}
	return passed, failed
}
