package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	l1app "github.com/sovereign-l1/l1/app"
	l1testutil "github.com/sovereign-l1/l1/tests/testutil"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	"github.com/sovereign-l1/l1/x/fee-collector/types"
)

func TestCollectFeesRecordsAccountingAndModuleBalance(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	collector := app.FeeCollectorKeeper
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(1_000)))[0]

	require.NoError(t, collector.CollectFeesFromAccount(ctx, user, sdk.NewCoins(coin(100)), types.FeeTypeGas))

	balances, err := collector.GetFeeBalances(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(coin(100)), balances.GasFees)
	require.Equal(t, sdk.NewCoins(coin(100)), balances.TotalCollected)
	require.Equal(t, sdk.NewCoins(coin(100)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.CollectorModuleName)))
	require.NoError(t, collector.AssertModuleAccountingInvariant(ctx))
}

func TestDistributeFeesRoutesTreasuryProtectionValidatorsAndBurn(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	collector := app.FeeCollectorKeeper
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(2_000)))[0]
	supplyBefore := app.BankKeeper.GetSupply(ctx, types.BaseDenom)
	validatorsBefore := app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName))

	require.NoError(t, collector.CollectFeesFromAccount(ctx, user, sdk.NewCoins(coin(1_000)), types.FeeTypeProtocol))
	history, err := collector.DistributeFees(ctx, 7)
	require.NoError(t, err)

	require.Equal(t, uint64(7), history.Epoch)
	require.Equal(t, sdk.NewCoins(coin(400)), history.Treasury)
	require.Equal(t, sdk.NewCoins(coin(200)), history.Protection)
	require.Equal(t, sdk.NewCoins(coin(380)), history.Validators)
	require.Equal(t, sdk.NewCoins(coin(20)), history.Burn)

	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.CollectorModuleName)))
	require.Equal(t, sdk.NewCoins(coin(400)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.TreasuryModuleName)))
	require.Equal(t, sdk.NewCoins(coin(200)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.ProtectionModuleName)))
	require.Equal(t, validatorsBefore.Add(coin(380)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	require.Equal(t, supplyBefore.Amount.Sub(sdkmath.NewInt(20)), app.BankKeeper.GetSupply(ctx, types.BaseDenom).Amount)
	require.NoError(t, collector.AssertModuleAccountingInvariant(ctx))
}

func TestRoundingRemainderIsDeterministicAndCannotCreateCoins(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	collector := app.FeeCollectorKeeper
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(100)))[0]

	params := types.DefaultParams()
	params.TreasuryBps = 3_333
	params.ProtectionBps = 3_333
	params.ValidatorsBps = 3_333
	params.BurnBps = 1
	msgServer := feecollectorkeeper.NewMsgServerImpl(collector)
	_, err := msgServer.UpdateFeeDistributionParams(ctx, &types.MsgUpdateFeeDistributionParams{
		Authority: collector.Authority(),
		Params:    params,
	})
	require.NoError(t, err)

	require.NoError(t, collector.CollectFeesFromAccount(ctx, user, sdk.NewCoins(coin(10)), types.FeeTypeForwarding))
	history, err := collector.DistributeFees(ctx, 9)
	require.NoError(t, err)

	require.Equal(t, sdk.NewCoins(coin(10)), history.Collected)
	require.Equal(t, sdk.NewCoins(coin(4)), history.Treasury)
	require.Equal(t, sdk.NewCoins(coin(3)), history.Protection)
	require.Equal(t, sdk.NewCoins(coin(3)), history.Validators)
	require.True(t, history.Burn.Empty())
	require.Equal(t, sdk.NewCoins(coin(1)), history.RoundingRemainder)
	require.Equal(t, history.Collected, history.Treasury.Add(history.Protection...).Add(history.Validators...).Add(history.Burn...))
}

func TestWrongDenomRejectedWithoutStateMutation(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	collector := app.FeeCollectorKeeper
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(100), sdk.NewInt64Coin(l1testutil.TestAssetDenom, 100)))[0]

	err := collector.CollectFeesFromAccount(ctx, user, sdk.NewCoins(sdk.NewInt64Coin(l1testutil.TestAssetDenom, 1)), types.FeeTypeGas)
	require.ErrorIs(t, err, types.ErrInvalidFee)

	balances, getErr := collector.GetFeeBalances(ctx)
	require.NoError(t, getErr)
	require.True(t, balances.AccountingBalance().Empty())
	require.True(t, balances.TotalCollected.Empty())
	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.CollectorModuleName)))
}

func TestMsgAuthorityAndDistributionParamSecurity(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feecollectorkeeper.NewMsgServerImpl(app.FeeCollectorKeeper)

	params := types.DefaultParams()
	params.BurnBps = 201
	_, err := msgServer.UpdateFeeDistributionParams(ctx, &types.MsgUpdateFeeDistributionParams{
		Authority: app.FeeCollectorKeeper.Authority(),
		Params:    params,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)

	_, err = msgServer.DistributeFees(ctx, &types.MsgDistributeFees{Authority: "ae1notgov", Epoch: 1})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestExportImportPreservesBalancesAndPendingDistribution(t *testing.T) {
	source := l1app.Setup(t, false)
	sourceCtx := source.NewContext(false)
	user := l1app.AddTestAddrsWithCoins(t, source, sourceCtx, 1, sdk.NewCoins(coin(1_000)))[0]
	require.NoError(t, source.FeeCollectorKeeper.CollectFeesFromAccount(sourceCtx, user, sdk.NewCoins(coin(123)), types.FeeTypeGas))

	exported, err := source.FeeCollectorKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.BankKeeper.MintCoins(targetCtx, minttypes.ModuleName, sdk.NewCoins(coin(123))))
	require.NoError(t, target.BankKeeper.SendCoinsFromModuleToModule(targetCtx, minttypes.ModuleName, types.CollectorModuleName, sdk.NewCoins(coin(123))))
	require.NoError(t, target.FeeCollectorKeeper.InitGenesis(targetCtx, *exported))

	imported, err := target.FeeCollectorKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
	require.NoError(t, target.FeeCollectorKeeper.AssertModuleAccountingInvariant(targetCtx))
}

func TestInvariantDetectsBankAndAccountingMismatch(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	collector := app.FeeCollectorKeeper
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(100)))[0]

	require.NoError(t, collector.CollectFeesFromAccount(ctx, user, sdk.NewCoins(coin(10)), types.FeeTypeGas))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(coin(1))))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.CollectorModuleName, sdk.NewCoins(coin(1))))

	err := collector.AssertModuleAccountingInvariant(ctx)
	require.ErrorIs(t, err, types.ErrAccounting)
	require.Contains(t, err.Error(), "module bank balance")
}

func coin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(types.BaseDenom, amount)
}
