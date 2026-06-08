package app

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestAppInvariantRegistryIncludesEveryRequiredInvariant(t *testing.T) {
	app := Setup(t, false)
	registry := app.AppInvariantRegistry()

	require.NoError(t, ValidateAppInvariantRegistry(registry))
	require.Len(t, registry, len(RequiredAppInvariantIDs()))

	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		seen[invariant.ID] = struct{}{}
		require.NotEmpty(t, invariant.Description)
		require.NotNil(t, invariant.Check)
	}
	for _, id := range RequiredAppInvariantIDs() {
		_, found := seen[id]
		require.Truef(t, found, "missing app invariant %s", id)
	}

	for _, id := range []string{
		nativeaccounttypes.InvariantModuleBankAccountingConsistent,
		nativeaccounttypes.InvariantRewardsCannotExceedAllocation,
		nativeaccounttypes.InvariantProtocolCriticalStateNotFrozenByRent,
		nativeaccounttypes.InvariantMinPoolDepositEnforced,
		nativeaccounttypes.InvariantMaxValidatorCountEnforced,
		nativeaccounttypes.InvariantMinValidatorStakeEnforced,
		nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected,
	} {
		_, found := seen[id]
		require.Truef(t, found, "UPDATE.md runtime invariant %s must be registered", id)
	}
}

func TestAppRuntimeInvariantsPassDefaultState(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	require.Empty(t, app.RunAppInvariants(ctx))
}

func TestAppBankAccountingInvariantDetectsSupplyMismatch(t *testing.T) {
	genesis := &banktypes.GenesisState{
		Balances: []banktypes.Balance{{
			Address: "AE1111111111111111111111111111111111111111111111111111",
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10)),
		}},
		Supply: sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 9)),
	}

	require.ErrorContains(t, validateBankSupplyMatchesBalances(genesis), "bank supply mismatch")
}

func TestAppDirectDelegationInvariantDetectsEnabledPolicy(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.RunAppInvariant(ctx, nativeaccounttypes.InvariantDirectUserValidatorDelegationRejected))

	params := nominatorpooltypes.DefaultParams()
	params.DirectUserDelegationEnabled = true

	require.ErrorContains(t, validateDirectUserDelegationPolicy(params), "direct user validator delegation is enabled")
}

func TestAppInvariantSecretFailureFixtureRejectsPrivateFields(t *testing.T) {
	input := nativeaccounttypes.NativeAccountInvariantInput{
		ExportedPayloads:             []string{"event.private_key=bad"},
		AEAddressRoundtripStable:     true,
		RawAddressRoundtripStable:    true,
		ActivationAttempts:           map[string]uint64{},
		TotalSupply:                  1,
		RewardBudget:                 1,
		MaxValidatorCount:            1,
		MinValidatorStake:            1,
		MinPoolDeposit:               1,
		ExportImportStable:           true,
		SystemReserveRunwayBlocks:    1,
		MinSystemReserveRunwayBlocks: 1,
		SystemTopUpOrder:             []string{"system_rent_top_up", "user_freeze_processing"},
		ProtocolCriticalExecutable:   true,
	}

	require.ErrorContains(t,
		nativeaccounttypes.RunNativeAccountInvariant(nativeaccounttypes.InvariantNoPrivateKeyOnChain, input),
		"secret-like payload rejected",
	)
}

func TestAppInvariantRegistryRejectsMissingRuntimeCheck(t *testing.T) {
	registry := []AppInvariant{{
		ID:          nativeaccounttypes.InvariantModuleBankAccountingConsistent,
		Description: "bank accounting",
	}}

	require.ErrorContains(t, ValidateAppInvariantRegistry(registry), "check are required")
}

func TestAppValidatorEntryInvariantPassesFundedValidator(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	_, validator := createFundedValidator(t, app, ctx, "app-invariant-validator-entry", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	require.True(t, validator.Tokens.GTE(validator.MinSelfDelegation))
	require.NoError(t, app.RunAppInvariant(ctx, nativeaccounttypes.InvariantMinValidatorStakeEnforced))
}

func TestSafeUint64UsesFallbackOnOverflow(t *testing.T) {
	require.Equal(t, uint64(7), safeUint64(sdkmath.NewInt(7), 99))
	require.Equal(t, uint64(99), safeUint64(sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 80)), 99))
}
