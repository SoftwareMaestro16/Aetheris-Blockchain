package params

import (
	"testing"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
)

func TestAetraMintPolicyMatchesEconomicsSpec(t *testing.T) {
	params := AetraMintParams()

	require.Equal(t, BaseDenom, params.MintDenom)
	require.Equal(t, BpsToLegacyDec(DefaultResponsivenessBps), params.InflationRateChange)
	require.Equal(t, BpsToLegacyDec(MinInflationBps), params.InflationMin)
	require.Equal(t, BpsToLegacyDec(MaxInflationBps), params.InflationMax)
	require.Equal(t, BpsToLegacyDec(DefaultTargetStakeBps), params.GoalBonded)
	require.Equal(t, int64(150), MinInflationBps)
	require.Equal(t, int64(500), MaxInflationBps)
	require.Equal(t, int64(6_000), DefaultTargetStakeBps)
	require.True(t, params.MaxSupply.IsZero(), "zero max supply means uncapped issuance")
	require.NoError(t, params.Validate())

	minter := AetraInitialMinter()
	require.Equal(t, BpsToLegacyDec(DefaultTargetInflationBps), minter.Inflation)
	require.NoError(t, minttypes.ValidateGenesis(*AetraMintGenesisState()))
}
