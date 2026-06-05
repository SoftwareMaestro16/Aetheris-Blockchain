package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestValidatorIncomeModelIsDeterministicAndBounded(t *testing.T) {
	income, err := ComputeValidatorIncome(ValidatorIncomeInput{
		TotalMintRewards: sdkmath.NewInt(1_000_000),
		TotalFeeRewards:  sdkmath.NewInt(100_000),
		ValidatorPower:   sdkmath.NewInt(20),
		TotalPower:       sdkmath.NewInt(100),
		CommissionBps:    1_000,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2_000), income.RewardWeightBps)
	require.Equal(t, sdkmath.NewInt(200_000), income.MintRewardShare)
	require.Equal(t, sdkmath.NewInt(20_000), income.FeeRewardShare)
	require.Equal(t, sdkmath.NewInt(22_000), income.ValidatorCommission)
	require.Equal(t, sdkmath.NewInt(242_000), income.ValidatorIncome)
	require.Equal(t, sdkmath.NewInt(198_000), income.DelegatorIncome)
}

func TestValidatorIncomeRejectsUnsafeCommissionAndPower(t *testing.T) {
	_, err := ComputeValidatorIncome(ValidatorIncomeInput{
		TotalMintRewards: sdkmath.NewInt(1),
		TotalFeeRewards:  sdkmath.NewInt(1),
		ValidatorPower:   sdkmath.NewInt(101),
		TotalPower:       sdkmath.NewInt(100),
		CommissionBps:    1_000,
	})
	require.ErrorContains(t, err, "validator power must be <= total power")

	require.Error(t, ValidateCommissionBounds(MinCommissionBps-1, 0))
	require.Error(t, ValidateCommissionBounds(MaxCommissionBps+1, 0))
	require.Error(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps+1))
	require.NoError(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps))
}

func TestBalanceControllerRaisesAndLowersInflationWithStaking(t *testing.T) {
	lowStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       4_000,
		BlockLoadBps:        DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Greater(t, lowStake.InflationBps, DefaultTargetInflationBps)

	highStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       9_000,
		BlockLoadBps:        DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Less(t, highStake.InflationBps, DefaultTargetInflationBps)
}

func TestBalanceControllerClampsInflationAndBurn(t *testing.T) {
	minOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MinInflationBps,
		StakeRatioBps:       BasisPoints,
		BlockLoadBps:        0,
	})
	require.NoError(t, err)
	require.Equal(t, MinInflationBps, minOut.InflationBps)

	maxOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MaxInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        BasisPoints,
		AnnualMint:          sdkmath.NewInt(100),
		AnnualBurn:          sdkmath.NewInt(200),
		AsyncQueueDepth:     10,
		FailedTxRateBps:     1_500,
	})
	require.NoError(t, err)
	require.Equal(t, MaxInflationBps, maxOut.InflationBps)
	require.Equal(t, int64(3_500), maxOut.BurnRatioBps)
	require.Equal(t, int64(5_500), maxOut.ValidatorFeeRatioBps)
	require.True(t, maxOut.Congested)
	require.True(t, maxOut.DeflationGuardActive)
	require.True(t, maxOut.QueueLimited)
	require.True(t, maxOut.RateLimited)
}

func TestBalanceControllerRejectsInvalidInputs(t *testing.T) {
	_, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MaxInflationBps + 1,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
	})
	require.ErrorContains(t, err, "current_inflation_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       -1,
		BlockLoadBps:        0,
	})
	require.ErrorContains(t, err, "stake_ratio_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
		AnnualMint:          sdkmath.NewInt(-1),
	})
	require.ErrorContains(t, err, "annual mint and burn")
}
