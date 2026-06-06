package params

import (
	sdkmath "cosmossdk.io/math"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func BpsToLegacyDec(bps int64) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(bps).Quo(sdkmath.LegacyNewDec(BasisPoints))
}

func AetherisInitialMinter() minttypes.Minter {
	return minttypes.InitialMinter(BpsToLegacyDec(DefaultTargetInflationBps))
}

func AetherisMintParams() minttypes.Params {
	params := minttypes.DefaultParams()
	params.MintDenom = BaseDenom
	params.InflationRateChange = BpsToLegacyDec(DefaultResponsivenessBps)
	params.InflationMin = BpsToLegacyDec(MinInflationBps)
	params.InflationMax = BpsToLegacyDec(MaxInflationBps)
	params.GoalBonded = BpsToLegacyDec(DefaultTargetStakeBps)
	params.MaxSupply = sdkmath.ZeroInt()
	return params
}

func AetherisMintGenesisState() *minttypes.GenesisState {
	return minttypes.NewGenesisState(AetherisInitialMinter(), AetherisMintParams())
}
