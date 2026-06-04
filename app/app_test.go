package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestOrbitalisChainConstants(t *testing.T) {
	require.Equal(t, "Orbitalis", appName)
	require.Equal(t, "orb", AccountAddressPrefix)
	require.Equal(t, "orbvaloper", ValidatorAddressPrefix)
	require.Equal(t, "orbvalcons", ConsensusAddressPrefix)
	require.Equal(t, appparams.BaseDenom, BondDenom)
	require.Equal(t, appparams.BaseDenom, sdk.DefaultBondDenom)
	require.Equal(t, int64(1_000_000_000), appparams.BaseUnitsPerDisplay)
	require.True(t, strings.HasSuffix(DefaultNodeHome, ".orbitalis"), DefaultNodeHome)
}

func TestDefaultGenesisIncludesNativeTokenMetadata(t *testing.T) {
	app, genesis := setup(true, 5)

	var bankGenState banktypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)

	var native banktypes.Metadata
	for _, metadata := range bankGenState.DenomMetadata {
		if metadata.Base == appparams.BaseDenom {
			native = metadata
			break
		}
	}

	requireNativeTokenMetadata(t, native)
}

func requireNativeTokenMetadata(t *testing.T, native banktypes.Metadata) {
	t.Helper()

	require.NoError(t, native.Validate())
	require.Equal(t, appparams.BaseDenom, native.Base)
	require.Equal(t, appparams.DisplayDenom, native.Display)
	require.Equal(t, appparams.TokenSymbol, native.Symbol)
	require.Equal(t, appparams.TokenName, native.Name)
	requireDenomUnit(t, native, appparams.BaseDenom, 0)
	requireDenomUnit(t, native, appparams.DisplayDenom, appparams.DisplayDenomExponent)
}

func requireDenomUnit(t *testing.T, metadata banktypes.Metadata, denom string, exponent uint32) {
	t.Helper()

	for _, unit := range metadata.DenomUnits {
		if unit.Denom == denom {
			require.Equal(t, exponent, unit.Exponent)
			return
		}
	}
	require.Failf(t, "missing denom unit", "denom %s", denom)
}

func TestDefaultGenesisValidatesAndSetsCustomModuleDefaults(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, []string{appparams.BaseDenom}, feesGenState.Params.AllowedFeeDenoms)
	require.Equal(t, "0.98", feesGenState.Params.ValidatorRewardsRatio)
	require.Equal(t, "0.02", feesGenState.Params.CommunityPoolRatio)

	stakingGenState := stakingtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
	require.Equal(t, appparams.BaseDenom, stakingGenState.Params.BondDenom)

	var mintGenState minttypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenState)
	require.Equal(t, appparams.BaseDenom, mintGenState.Params.MintDenom)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
}

func TestNativeTokenRuntimeMetadataSendAndSupply(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	metadata, found := app.BankKeeper.GetDenomMetaData(ctx, appparams.BaseDenom)
	require.True(t, found)
	requireNativeTokenMetadata(t, metadata)

	addrs := AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	sender, recipient := addrs[0], addrs[1]
	beforeSupply := app.BankKeeper.GetSupply(ctx, appparams.BaseDenom)
	sendAmount := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 123_456))

	require.NoError(t, app.BankKeeper.SendCoins(ctx, sender, recipient, sendAmount))

	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 876_544), app.BankKeeper.GetBalance(ctx, sender, appparams.BaseDenom))
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 1_123_456), app.BankKeeper.GetBalance(ctx, recipient, appparams.BaseDenom))
	require.Equal(t, beforeSupply, app.BankKeeper.GetSupply(ctx, appparams.BaseDenom))
}

func TestCustomModuleGenesisInitExportRoundTrip(t *testing.T) {
	app := Setup(t, false)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, feestypes.DefaultGenesisState(), &feesGenState)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)
	require.NoError(t, tokenfactoryGenState.Validate())

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
	require.NoError(t, dexGenState.Validate())
}

func TestDefaultGenesisRejectsCorruptedPrototypeModuleState(t *testing.T) {
	app, baseGenesis := setup(true, 5)
	cdc := app.AppCodec()
	admin := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()

	tests := map[string]func(GenesisState){
		"invalid native metadata": func(genesis GenesisState) {
			var bankGenState banktypes.GenesisState
			cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)
			bankGenState.DenomMetadata = []banktypes.Metadata{{Base: "bad denom", Display: appparams.DisplayDenom}}
			genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(&bankGenState)
		},
		"invalid staking denom": func(genesis GenesisState) {
			stakingGenState := stakingtypes.GetGenesisStateFromAppState(cdc, genesis)
			stakingGenState.Params.BondDenom = "bad denom"
			genesis[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenState)
		},
		"invalid fees params": func(genesis GenesisState) {
			var feesGenState feestypes.GenesisState
			cdc.MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
			feesGenState.Params.AllowedFeeDenoms = []string{appparams.TestAssetDenom}
			genesis[feestypes.ModuleName] = cdc.MustMarshalJSON(&feesGenState)
		},
		"invalid tokenfactory metadata": func(genesis GenesisState) {
			tokenfactoryGenState := tokenfactorytypes.GenesisState{Denoms: []tokenfactorytypes.DenomAuthorityMetadata{{
				Denom: "factory/" + admin + "/" + appparams.BaseDenom,
				Admin: admin,
			}}}
			genesis[tokenfactorytypes.ModuleName] = cdc.MustMarshalJSON(&tokenfactoryGenState)
		},
		"duplicate dex pool pair": func(genesis GenesisState) {
			dexGenState := dextypes.GenesisState{NextPoolId: 3, Pools: []dextypes.Pool{
				{Id: 1, Denom0: "aaa", Denom1: appparams.BaseDenom, Reserve0: "1", Reserve1: "1", TotalShares: "1", LpDenom: "lp/1"},
				{Id: 2, Denom0: "aaa", Denom1: appparams.BaseDenom, Reserve0: "1", Reserve1: "1", TotalShares: "1", LpDenom: "lp/2"},
			}}
			genesis[dextypes.ModuleName] = cdc.MustMarshalJSON(&dexGenState)
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			genesis := cloneGenesisState(baseGenesis)
			mutate(genesis)
			err := app.BasicModuleManager.ValidateGenesis(cdc, app.TxConfig(), genesis)
			require.Error(t, err)
		})
	}
}

func TestPrototypeModuleAccountPermissionsAreNarrow(t *testing.T) {
	expected := map[string][]string{
		authtypes.FeeCollectorName:                  nil,
		distrtypes.ModuleName:                       nil,
		minttypes.ModuleName:                        {authtypes.Minter},
		stakingtypes.BondedPoolName:                 {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:              {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                         {authtypes.Burner},
		protocolpooltypes.ModuleName:                nil,
		protocolpooltypes.ProtocolPoolEscrowAccount: nil,
		tokenfactorytypes.ModuleName:                {authtypes.Minter, authtypes.Burner},
		dextypes.ModuleName:                         {authtypes.Minter, authtypes.Burner},
		feestypes.ModuleName:                        nil,
	}
	require.Equal(t, expected, GetMaccPerms())

	blocked := BlockedAddresses()
	for moduleName := range expected {
		addr := authtypes.NewModuleAddress(moduleName).String()
		if moduleName == govtypes.ModuleName {
			require.False(t, blocked[addr])
			continue
		}
		require.True(t, blocked[addr], moduleName)
	}
}

func cloneGenesisState(genesis GenesisState) GenesisState {
	clone := make(GenesisState, len(genesis))
	for moduleName, raw := range genesis {
		clone[moduleName] = append(json.RawMessage(nil), raw...)
	}
	return clone
}
