package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultParamsUseAetraBalancedFeeSplit(t *testing.T) {
	params := DefaultParams()

	require.NoError(t, params.Validate())
	require.Equal(t, uint32(5_000), params.BurnBps)
	require.Equal(t, uint32(3_500), params.ValidatorsBps)
	require.Equal(t, uint32(1_500), params.TreasuryBps)
	require.Equal(t, uint32(0), params.ProtectionBps)
}

func TestSplitFeesUsesBurnRewardTreasuryModel(t *testing.T) {
	distribution, remainder, err := SplitFees(DefaultParams(), sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 10_000)))

	require.NoError(t, err)
	require.True(t, remainder.Empty())
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 5_000)), distribution.Burn)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 3_500)), distribution.Validators)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 1_500)), distribution.Treasury)
	require.True(t, distribution.Protection.Empty())
}

func TestParamsRejectUnsafeFeeSplit(t *testing.T) {
	params := DefaultParams()
	params.BurnBps = 200
	params.TreasuryBps = 4_000
	params.ProtectionBps = 2_000
	params.ValidatorsBps = 3_800

	require.ErrorContains(t, params.Validate(), "burn_bps")

	params = DefaultParams()
	params.ProtectionBps = 100
	params.BurnBps -= 100
	require.ErrorContains(t, params.Validate(), "protection_bps")
}
