package app

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestPoSCreateValidatorWithNaet(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	selfDelegation := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	valAddr, validator := createFundedValidator(t, app, ctx, "phase4-create-validator", selfDelegation)

	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.Equal(t, selfDelegation, validator.Tokens)
	require.Equal(t, sdkmath.OneInt(), validator.MinSelfDelegation)
	require.Equal(t, int64(10), validator.GetConsensusPower(sdk.DefaultPowerReduction))

	delegation, err := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(valAddr), valAddr)
	require.NoError(t, err)
	require.Equal(t, validator.DelegatorShares, delegation.Shares)
}

func TestPoSDelegationUpdatesValidatorPower(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	validator := GetBondedTestValidator(t, app, ctx)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.False(t, validator.Jailed)

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	require.Equal(t, BondDenom, bondDenom)

	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	beforeTokens := validator.Tokens
	beforePower := validator.GetConsensusPower(sdk.DefaultPowerReduction)

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	))
	require.NoError(t, err)

	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)
	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	require.Equal(t, beforeTokens.Add(delegation), updatedValidator.Tokens)
	require.Equal(t, beforePower+5, updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction))

	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, updates)

	var found bool
	for _, update := range updates {
		if update.Power == updatedValidator.GetConsensusPower(sdk.DefaultPowerReduction) {
			found = true
			break
		}
	}
	require.True(t, found, "expected validator-set update with new voting power")
}

func TestPoSUnbondingFlow(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	validator := GetBondedTestValidator(t, app, ctx)
	delegation := sdk.TokensFromConsensusPower(4, sdk.DefaultPowerReduction)
	unbond := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err := msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	))
	require.NoError(t, err)
	balanceBeforeUnbond := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)

	undelegateRes, err := msgServer.Undelegate(ctx, stakingtypes.NewMsgUndelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, unbond),
	))
	require.NoError(t, err)
	require.True(t, undelegateRes.CompletionTime.After(ctx.BlockTime()))

	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)
	remaining, err := app.StakingKeeper.GetDelegation(ctx, delegator, valAddr)
	require.NoError(t, err)
	require.True(t, remaining.Shares.IsPositive())

	unbonding, err := app.StakingKeeper.GetUnbondingDelegation(ctx, delegator, valAddr)
	require.NoError(t, err)
	require.Len(t, unbonding.Entries, 1)
	require.Equal(t, unbond, unbonding.Entries[0].Balance)
	require.Equal(t, balanceBeforeUnbond, app.BankKeeper.GetBalance(ctx, delegator, BondDenom), "unbonded stake must not return before completion time")
}

func TestPoSRedelegationFlow(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	source := GetBondedTestValidator(t, app, ctx)
	dstValAddr, destination := createFundedValidator(t, app, ctx, "phase4-redelegate-dst", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))
	require.Equal(t, dstValAddr.String(), destination.OperatorAddress)

	delegation := sdk.TokensFromConsensusPower(4, sdk.DefaultPowerReduction)
	redelegate := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err := msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
		delegator.String(),
		source.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	))
	require.NoError(t, err)

	redelegateRes, err := msgServer.BeginRedelegate(ctx, stakingtypes.NewMsgBeginRedelegate(
		delegator.String(),
		source.OperatorAddress,
		destination.OperatorAddress,
		sdk.NewCoin(BondDenom, redelegate),
	))
	require.NoError(t, err)
	require.True(t, redelegateRes.CompletionTime.After(ctx.BlockTime()))

	srcValAddr, err := sdk.ValAddressFromBech32(source.OperatorAddress)
	require.NoError(t, err)
	sourceDelegation, err := app.StakingKeeper.GetDelegation(ctx, delegator, srcValAddr)
	require.NoError(t, err)
	require.True(t, sourceDelegation.Shares.IsPositive())

	destinationDelegation, err := app.StakingKeeper.GetDelegation(ctx, delegator, dstValAddr)
	require.NoError(t, err)
	require.True(t, destinationDelegation.Shares.IsPositive())

	storedRedelegation, err := app.StakingKeeper.GetRedelegation(ctx, delegator, srcValAddr, dstValAddr)
	require.NoError(t, err)
	require.Len(t, storedRedelegation.Entries, 1)
	require.Equal(t, redelegate, storedRedelegation.Entries[0].InitialBalance)
}

func TestPoSRejectsInvalidDelegations(t *testing.T) {
	tests := []struct {
		name             string
		fundedCoins      sdk.Coins
		delegatorAddress string
		delegationAmount sdk.Coin
		validatorAddress string
	}{
		{
			name:             "wrong denom",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10_000_000), sdk.NewInt64Coin("uatom", 10_000_000)),
			delegationAmount: sdk.NewInt64Coin("uatom", 5_000_000),
		},
		{
			name:             "insufficient funds",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)),
			delegationAmount: sdk.NewInt64Coin(BondDenom, 5_000_000),
		},
		{
			name:             "invalid validator address",
			fundedCoins:      sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10_000_000)),
			delegationAmount: sdk.NewInt64Coin(BondDenom, 5_000_000),
			validatorAddress: "not-a-validator-address",
		},
		{
			name:             "invalid delegator address",
			delegatorAddress: "not-a-delegator-address",
			delegationAmount: sdk.NewInt64Coin(BondDenom, 5_000_000),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := Setup(t, false)
			ctx := app.NewContext(false)
			validator := GetBondedTestValidator(t, app, ctx)
			validatorAddress := validator.OperatorAddress
			if tc.validatorAddress != "" {
				validatorAddress = tc.validatorAddress
			}

			delegatorAddress := tc.delegatorAddress
			if delegatorAddress == "" {
				delegator := AddTestAddrsWithCoins(t, app, ctx, 1, tc.fundedCoins)[0]
				delegatorAddress = delegator.String()
			}
			msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
			_, err := msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
				delegatorAddress,
				validatorAddress,
				tc.delegationAmount,
			))
			require.Error(t, err)
		})
	}
}

func TestSlashingParamsAndSigningInfoRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	params, err := app.SlashingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Positive(t, params.SignedBlocksWindow)
	require.True(t, params.MinSignedPerWindow.IsPositive())
	require.True(t, params.SlashFractionDoubleSign.IsPositive())
	require.True(t, params.SlashFractionDowntime.IsPositive())

	validator := GetBondedTestValidator(t, app, ctx)
	consAddrBytes, err := validator.GetConsAddr()
	require.NoError(t, err)
	consAddr := sdk.ConsAddress(consAddrBytes)
	expected := slashingtypes.NewValidatorSigningInfo(consAddr, 7, 3, time.Unix(0, 0).UTC(), false, 2)

	require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, expected))
	actual, err := app.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	require.NoError(t, app.SlashingKeeper.SetMissedBlockBitmapValue(ctx, consAddr, 5, true))
	missed, err := app.SlashingKeeper.GetMissedBlockBitmapValue(ctx, consAddr, 5)
	require.NoError(t, err)
	require.True(t, missed)

	require.NoError(t, app.SlashingKeeper.SetMissedBlockBitmapValue(ctx, consAddr, 5, false))
	missed, err = app.SlashingKeeper.GetMissedBlockBitmapValue(ctx, consAddr, 5)
	require.NoError(t, err)
	require.False(t, missed)
}

func TestStakingRewardsDistributionCanBeWithdrawn(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockTime(time.Now().UTC())

	validator := GetBondedTestValidator(t, app, ctx)
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	require.NoError(t, err)
	delegation := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	delegator := AddTestAddrsIncremental(app, ctx, 1, delegation.MulRaw(2))[0]

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(
		delegator.String(),
		validator.OperatorAddress,
		sdk.NewCoin(BondDenom, delegation),
	))
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(time.Second))
	updatedValidator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	distrMsgServer := distrkeeper.NewMsgServerImpl(app.DistrKeeper)
	depositor := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	_, err = distrMsgServer.DepositValidatorRewardsPool(ctx, distrtypes.NewMsgDepositValidatorRewardsPool(
		depositor.String(),
		updatedValidator.OperatorAddress,
		sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 100_000)),
	))
	require.NoError(t, err)

	balanceBefore := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)
	_, err = distrMsgServer.WithdrawDelegatorReward(ctx, distrtypes.NewMsgWithdrawDelegatorReward(
		delegator.String(),
		validator.OperatorAddress,
	))
	require.NoError(t, err)

	balanceAfter := app.BankKeeper.GetBalance(ctx, delegator, BondDenom)
	require.True(t, balanceAfter.Amount.GT(balanceBefore.Amount), "delegator must receive naet staking rewards")
}

func TestPoSMintPolicyIsNaetAndUncappedWithBoundedInflation(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	params, err := app.MintKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, BondDenom, params.MintDenom)
	require.True(t, params.MaxSupply.IsZero(), "zero max supply means uncapped PoS issuance in Cosmos SDK mint params")
	require.NoError(t, params.Validate())
	require.False(t, params.InflationRateChange.IsNegative())
	require.False(t, params.InflationMin.IsNegative())
	require.True(t, params.InflationMax.GTE(params.InflationMin))
	require.True(t, params.GoalBonded.IsPositive())
	require.True(t, params.GoalBonded.LTE(sdkmath.LegacyOneDec()))
	require.Positive(t, params.BlocksPerYear)

	defaults := minttypes.DefaultParams()
	require.Equal(t, defaults.InflationRateChange, params.InflationRateChange)
	require.Equal(t, defaults.InflationMin, params.InflationMin)
	require.Equal(t, defaults.InflationMax, params.InflationMax)
	require.Equal(t, defaults.GoalBonded, params.GoalBonded)
	require.Equal(t, defaults.BlocksPerYear, params.BlocksPerYear)
}

func TestAddTestAddrsUsesBondDenom(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	addr := AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(123))[0]
	require.Equal(t, sdk.NewInt64Coin(BondDenom, 123), app.BankKeeper.GetBalance(ctx, addr, BondDenom))
}

func createFundedValidator(t *testing.T, app *L1App, ctx sdk.Context, moniker string, selfDelegation sdkmath.Int) (sdk.ValAddress, stakingtypes.Validator) {
	t.Helper()
	operator := AddTestAddrsIncremental(app, ctx, 1, selfDelegation.MulRaw(2))[0]
	valAddr := sdk.ValAddress(operator)
	msg, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		ed25519.GenPrivKey().PubKey(),
		sdk.NewCoin(BondDenom, selfDelegation),
		stakingtypes.Description{Moniker: moniker},
		stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
		sdkmath.OneInt(),
	)
	require.NoError(t, err)

	msgServer := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(t, err)

	_, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)

	validator, err := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	return valAddr, validator
}
