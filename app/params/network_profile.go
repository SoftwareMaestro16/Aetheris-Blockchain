package params

import "fmt"

const (
	AetraConsensusEngine = "CometBFT BFT"
	AetraStakingModel    = "PoS + delegation + nomination pools"
	AetraPrimaryVM       = "CosmWasm"
	AetraOptionalVM      = "EVM later"
	AetraHardwareProfile = "medium"

	AetraValidatorSetMin            = 100
	AetraValidatorSetGenesisMin     = 100
	AetraValidatorSetGenesisMax     = 128
	AetraValidatorSetGrowthMin      = 150
	AetraValidatorSetGrowthMax      = 200
	AetraValidatorSetMatureMin      = 250
	AetraValidatorSetMatureMax      = 300
	AetraValidatorSetMax            = 300
	AetraBlockTimeMinSeconds        = 5
	AetraBlockTimeMaxSeconds        = 8
	AetraNormalFinalityMinSeconds   = 5
	AetraNormalFinalityMaxSeconds   = 15
	AetraStressFinalityMinSeconds   = 20
	AetraStressFinalityMaxSeconds   = 90
	AetraWorstFinalityTargetSeconds = 120
	AetraNormalInflationMinBps      = int64(200)
	AetraNormalInflationMaxBps      = int64(500)
	AetraDelegatorAPRTargetMinBps   = int64(500)
	AetraDelegatorAPRTargetMaxBps   = int64(800)
	AetraFeeBurnShareMinBps         = int64(3_000)
	AetraFeeBurnShareMaxBps         = int64(6_000)
	AetraFeeRewardShareMinBps       = int64(2_000)
	AetraFeeRewardShareMaxBps       = int64(4_000)
	AetraFeeTreasuryShareMinBps     = int64(1_000)
	AetraFeeTreasuryShareMaxBps     = int64(2_000)
)

type NetworkProfile struct {
	ConsensusEngine            string
	StakingModel               string
	PrimaryVM                  string
	OptionalVM                 string
	HardwareProfile            string
	ValidatorSetMin            int
	ValidatorSetGenesisMin     int
	ValidatorSetGenesisMax     int
	ValidatorSetGrowthMin      int
	ValidatorSetGrowthMax      int
	ValidatorSetMatureMin      int
	ValidatorSetMatureMax      int
	ValidatorSetMax            int
	BlockTimeMinSeconds        int
	BlockTimeMaxSeconds        int
	NormalFinalityMinSeconds   int
	NormalFinalityMaxSeconds   int
	StressFinalityMinSeconds   int
	StressFinalityMaxSeconds   int
	WorstFinalityTargetSeconds int
	NormalInflationMinBps      int64
	NormalInflationMaxBps      int64
	DelegatorAPRTargetMinBps   int64
	DelegatorAPRTargetMaxBps   int64
	FeeBurnShareMinBps         int64
	FeeBurnShareMaxBps         int64
	FeeRewardShareMinBps       int64
	FeeRewardShareMaxBps       int64
	FeeTreasuryShareMinBps     int64
	FeeTreasuryShareMaxBps     int64
}

func DefaultNetworkProfile() NetworkProfile {
	return NetworkProfile{
		ConsensusEngine:            AetraConsensusEngine,
		StakingModel:               AetraStakingModel,
		PrimaryVM:                  AetraPrimaryVM,
		OptionalVM:                 AetraOptionalVM,
		HardwareProfile:            AetraHardwareProfile,
		ValidatorSetMin:            AetraValidatorSetMin,
		ValidatorSetGenesisMin:     AetraValidatorSetGenesisMin,
		ValidatorSetGenesisMax:     AetraValidatorSetGenesisMax,
		ValidatorSetGrowthMin:      AetraValidatorSetGrowthMin,
		ValidatorSetGrowthMax:      AetraValidatorSetGrowthMax,
		ValidatorSetMatureMin:      AetraValidatorSetMatureMin,
		ValidatorSetMatureMax:      AetraValidatorSetMatureMax,
		ValidatorSetMax:            AetraValidatorSetMax,
		BlockTimeMinSeconds:        AetraBlockTimeMinSeconds,
		BlockTimeMaxSeconds:        AetraBlockTimeMaxSeconds,
		NormalFinalityMinSeconds:   AetraNormalFinalityMinSeconds,
		NormalFinalityMaxSeconds:   AetraNormalFinalityMaxSeconds,
		StressFinalityMinSeconds:   AetraStressFinalityMinSeconds,
		StressFinalityMaxSeconds:   AetraStressFinalityMaxSeconds,
		WorstFinalityTargetSeconds: AetraWorstFinalityTargetSeconds,
		NormalInflationMinBps:      AetraNormalInflationMinBps,
		NormalInflationMaxBps:      AetraNormalInflationMaxBps,
		DelegatorAPRTargetMinBps:   AetraDelegatorAPRTargetMinBps,
		DelegatorAPRTargetMaxBps:   AetraDelegatorAPRTargetMaxBps,
		FeeBurnShareMinBps:         AetraFeeBurnShareMinBps,
		FeeBurnShareMaxBps:         AetraFeeBurnShareMaxBps,
		FeeRewardShareMinBps:       AetraFeeRewardShareMinBps,
		FeeRewardShareMaxBps:       AetraFeeRewardShareMaxBps,
		FeeTreasuryShareMinBps:     AetraFeeTreasuryShareMinBps,
		FeeTreasuryShareMaxBps:     AetraFeeTreasuryShareMaxBps,
	}
}

func (p NetworkProfile) Validate() error {
	if p.ConsensusEngine != AetraConsensusEngine {
		return fmt.Errorf("consensus engine must be %q", AetraConsensusEngine)
	}
	if p.StakingModel != AetraStakingModel {
		return fmt.Errorf("staking model must be %q", AetraStakingModel)
	}
	if p.PrimaryVM != AetraPrimaryVM {
		return fmt.Errorf("primary VM must be %q", AetraPrimaryVM)
	}
	if p.ValidatorSetMin < 100 || p.ValidatorSetMax > 300 || p.ValidatorSetMin > p.ValidatorSetMax {
		return fmt.Errorf("validator set must stay within 100-300 active validators")
	}
	if p.ValidatorSetGenesisMin < p.ValidatorSetMin || p.ValidatorSetGenesisMax > 128 || p.ValidatorSetGenesisMin > p.ValidatorSetGenesisMax {
		return fmt.Errorf("genesis validator set must stay within 100-128 active validators")
	}
	if p.ValidatorSetGrowthMin < 150 || p.ValidatorSetGrowthMax > 200 || p.ValidatorSetGrowthMin > p.ValidatorSetGrowthMax {
		return fmt.Errorf("growth validator set must stay within 150-200 active validators")
	}
	if p.ValidatorSetMatureMin < 250 || p.ValidatorSetMatureMax > p.ValidatorSetMax || p.ValidatorSetMatureMin > p.ValidatorSetMatureMax {
		return fmt.Errorf("mature validator set must stay within 250-300 active validators")
	}
	if p.BlockTimeMinSeconds < 5 || p.BlockTimeMaxSeconds > 8 || p.BlockTimeMinSeconds > p.BlockTimeMaxSeconds {
		return fmt.Errorf("block time must stay within 5-8 seconds")
	}
	if p.NormalFinalityMinSeconds < p.BlockTimeMinSeconds || p.NormalFinalityMaxSeconds > 15 || p.NormalFinalityMinSeconds > p.NormalFinalityMaxSeconds {
		return fmt.Errorf("normal finality must stay within 5-15 seconds")
	}
	if p.StressFinalityMinSeconds < 20 || p.StressFinalityMaxSeconds > 90 || p.StressFinalityMinSeconds > p.StressFinalityMaxSeconds {
		return fmt.Errorf("stress finality must stay within 20-90 seconds")
	}
	if p.WorstFinalityTargetSeconds > 120 || p.WorstFinalityTargetSeconds < p.StressFinalityMaxSeconds {
		return fmt.Errorf("worst finality target must be <= 120 seconds and cover stress finality")
	}
	if err := validateBpsRange("normal_inflation", p.NormalInflationMinBps, p.NormalInflationMaxBps, 200, 500); err != nil {
		return err
	}
	if err := validateBpsRange("delegator_apr_target", p.DelegatorAPRTargetMinBps, p.DelegatorAPRTargetMaxBps, 500, 800); err != nil {
		return err
	}
	if err := validateBpsRange("fee_burn_share", p.FeeBurnShareMinBps, p.FeeBurnShareMaxBps, 3_000, 6_000); err != nil {
		return err
	}
	if err := validateBpsRange("fee_reward_share", p.FeeRewardShareMinBps, p.FeeRewardShareMaxBps, 2_000, 4_000); err != nil {
		return err
	}
	if err := validateBpsRange("fee_treasury_share", p.FeeTreasuryShareMinBps, p.FeeTreasuryShareMaxBps, 1_000, 2_000); err != nil {
		return err
	}
	if DefaultTargetInflationBps < p.NormalInflationMinBps || DefaultTargetInflationBps > p.NormalInflationMaxBps {
		return fmt.Errorf("default target inflation must remain inside normal inflation range")
	}
	return nil
}

func validateBpsRange(name string, min, max, allowedMin, allowedMax int64) error {
	if min < allowedMin || max > allowedMax || min > max {
		return fmt.Errorf("%s must stay within %d-%d bps", name, allowedMin, allowedMax)
	}
	return nil
}
