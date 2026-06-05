package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	BasisPoints = int64(10_000)

	MinCommissionBps            = int64(100)
	MaxCommissionBps            = int64(2_000)
	MaxDailyCommissionChangeBps = int64(100)

	MinInflationBps             = int64(100)
	MaxInflationBps             = int64(500)
	DefaultTargetInflationBps   = int64(300)
	DefaultTargetStakeBps       = int64(6_700)
	DefaultResponsivenessBps    = int64(800)
	NormalBurnRatioBps          = int64(3_000)
	CongestedBurnRatioBps       = int64(4_000)
	MinBurnRatioBps             = int64(1_000)
	MaxBurnRatioBps             = int64(5_000)
	TreasuryFeeRatioBps         = int64(1_000)
	DefaultTargetLoadBps        = int64(7_000)
	HighCongestionLoadBps       = int64(9_000)
	DefaultMaxLoadMultiplierBps = int64(40_000)
)

type ValidatorIncomeInput struct {
	TotalMintRewards sdkmath.Int
	TotalFeeRewards  sdkmath.Int
	ValidatorPower   sdkmath.Int
	TotalPower       sdkmath.Int
	CommissionBps    int64
}

type ValidatorIncome struct {
	RewardWeightBps       int64
	MintRewardShare       sdkmath.Int
	FeeRewardShare        sdkmath.Int
	ValidatorCommission   sdkmath.Int
	ValidatorIncome       sdkmath.Int
	DelegatorIncome       sdkmath.Int
	DelegatorGrossRewards sdkmath.Int
}

type BalanceControllerInput struct {
	CurrentInflationBps int64
	StakeRatioBps       int64
	BlockLoadBps        int64
	AnnualMint          sdkmath.Int
	AnnualBurn          sdkmath.Int
	AsyncQueueDepth     uint64
	FailedTxRateBps     int64
}

type BalanceControllerOutput struct {
	InflationBps         int64
	BurnRatioBps         int64
	ValidatorFeeRatioBps int64
	Congested            bool
	DeflationGuardActive bool
	QueueLimited         bool
	RateLimited          bool
}

func ComputeValidatorIncome(input ValidatorIncomeInput) (ValidatorIncome, error) {
	totalMintRewards := normalizeInt(input.TotalMintRewards)
	totalFeeRewards := normalizeInt(input.TotalFeeRewards)
	validatorPower := normalizeInt(input.ValidatorPower)
	totalPower := normalizeInt(input.TotalPower)
	if totalMintRewards.IsNegative() || totalFeeRewards.IsNegative() {
		return ValidatorIncome{}, fmt.Errorf("validator rewards must not be negative")
	}
	if !validatorPower.IsPositive() || !totalPower.IsPositive() {
		return ValidatorIncome{}, fmt.Errorf("validator and total power must be positive")
	}
	if validatorPower.GT(totalPower) {
		return ValidatorIncome{}, fmt.Errorf("validator power must be <= total power")
	}
	if err := ValidateCommissionBounds(input.CommissionBps, 0); err != nil {
		return ValidatorIncome{}, err
	}

	rewardWeightBps := validatorPower.MulRaw(BasisPoints).Quo(totalPower).Int64()
	mintShare := ProportionalShare(totalMintRewards, validatorPower, totalPower)
	feeShare := ProportionalShare(totalFeeRewards, validatorPower, totalPower)
	gross := mintShare.Add(feeShare)
	commission := ApplyBps(gross, input.CommissionBps)
	delegatorIncome := gross.Sub(commission)

	return ValidatorIncome{
		RewardWeightBps:       rewardWeightBps,
		MintRewardShare:       mintShare,
		FeeRewardShare:        feeShare,
		ValidatorCommission:   commission,
		ValidatorIncome:       gross.Add(commission),
		DelegatorIncome:       delegatorIncome,
		DelegatorGrossRewards: gross,
	}, nil
}

func BalanceController(input BalanceControllerInput) (BalanceControllerOutput, error) {
	if err := validateBps("current_inflation_bps", input.CurrentInflationBps, MinInflationBps, MaxInflationBps); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("stake_ratio_bps", input.StakeRatioBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("block_load_bps", input.BlockLoadBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	if err := validateBps("failed_tx_rate_bps", input.FailedTxRateBps, 0, BasisPoints); err != nil {
		return BalanceControllerOutput{}, err
	}
	annualMint := normalizeInt(input.AnnualMint)
	annualBurn := normalizeInt(input.AnnualBurn)
	if annualMint.IsNegative() || annualBurn.IsNegative() {
		return BalanceControllerOutput{}, fmt.Errorf("annual mint and burn must not be negative")
	}

	inflationDelta := DefaultResponsivenessBps * (DefaultTargetStakeBps - input.StakeRatioBps) / BasisPoints
	inflation := clampInt64(input.CurrentInflationBps+inflationDelta, MinInflationBps, MaxInflationBps)

	burnRatio := NormalBurnRatioBps
	congested := input.BlockLoadBps > HighCongestionLoadBps
	if congested {
		burnRatio = CongestedBurnRatioBps
	}

	deflationGuard := false
	if annualMint.IsPositive() && annualBurn.GT(annualMint.MulRaw(125).QuoRaw(100)) {
		deflationGuard = true
		burnRatio = clampInt64(burnRatio-500, MinBurnRatioBps, MaxBurnRatioBps)
	}
	burnRatio = clampInt64(burnRatio, MinBurnRatioBps, MaxBurnRatioBps)
	validatorFeeRatio := BasisPoints - burnRatio - TreasuryFeeRatioBps
	if validatorFeeRatio < 0 {
		return BalanceControllerOutput{}, fmt.Errorf("burn and treasury ratios exceed 100%%")
	}

	return BalanceControllerOutput{
		InflationBps:         inflation,
		BurnRatioBps:         burnRatio,
		ValidatorFeeRatioBps: validatorFeeRatio,
		Congested:            congested,
		DeflationGuardActive: deflationGuard,
		QueueLimited:         input.AsyncQueueDepth > 0 && input.BlockLoadBps > DefaultTargetLoadBps,
		RateLimited:          input.FailedTxRateBps > 1_000 || input.BlockLoadBps > HighCongestionLoadBps,
	}, nil
}

func ValidateCommissionBounds(commissionBps, dailyChangeBps int64) error {
	if err := validateBps("commission_bps", commissionBps, MinCommissionBps, MaxCommissionBps); err != nil {
		return err
	}
	return validateBps("daily_commission_change_bps", dailyChangeBps, 0, MaxDailyCommissionChangeBps)
}

func ProportionalShare(total, numerator, denominator sdkmath.Int) sdkmath.Int {
	total = normalizeInt(total)
	numerator = normalizeInt(numerator)
	denominator = normalizeInt(denominator)
	if total.IsZero() || numerator.IsZero() || !denominator.IsPositive() {
		return sdkmath.ZeroInt()
	}
	return total.Mul(numerator).Quo(denominator)
}

func ApplyBps(amount sdkmath.Int, bps int64) sdkmath.Int {
	amount = normalizeInt(amount)
	if amount.IsZero() || bps == 0 {
		return sdkmath.ZeroInt()
	}
	return amount.MulRaw(bps).QuoRaw(BasisPoints)
}

func normalizeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func validateBps(name string, value, min, max int64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return nil
}

func clampInt64(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
