package integration_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	testutil "github.com/sovereign-l1/l1/tests/testutil"
	dexkeeper "github.com/sovereign-l1/l1/x/dex/keeper"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tfkeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	tftypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestSignedBankTxReplayIsRejectedAfterSequenceIncrement(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-1")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	_, sequenceBefore := testutil.AccountNumberAndSequence(t, app, ctx, sender)
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), 200_000)

	first := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, first.TxResults, 1)
	require.Zero(t, first.TxResults[0].Code, first.TxResults[0].Log)
	testutil.Commit(t, app)

	ctxAfterFirst := testutil.NewContext(app, 2)
	recipientAfterFirst := app.BankKeeper.GetBalance(ctxAfterFirst, recipient, "naet")
	require.Equal(t, sdkmath.NewInt(1_000_100), recipientAfterFirst.Amount)
	_, sequenceAfterFirst := testutil.AccountNumberAndSequence(t, app, ctxAfterFirst, sender)
	require.Equal(t, sequenceBefore+1, sequenceAfterFirst)

	replay := testutil.FinalizeBlock(t, app, 2, txBytes)
	require.Len(t, replay.TxResults, 1)
	require.NotZero(t, replay.TxResults[0].Code, "replayed tx with stale sequence must fail")
	testutil.Commit(t, app)

	ctxAfterReplay := testutil.NewContext(app, 3)
	require.Equal(t, recipientAfterFirst, app.BankKeeper.GetBalance(ctxAfterReplay, recipient, "naet"))
	_, sequenceAfterReplay := testutil.AccountNumberAndSequence(t, app, ctxAfterReplay, sender)
	require.Equal(t, sequenceAfterFirst, sequenceAfterReplay)
}

func TestInvalidSignerTxFailsBeforeBalanceMutation(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-2")
	ctx := testutil.NewContext(app, 1)
	_, victim := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	attackerPriv, _ := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	before := app.BankKeeper.GetBalance(ctx, recipient, "naet")

	msg := banktypes.NewMsgSend(victim, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, attackerPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "tx signed by non-msg signer must fail")

	after := app.BankKeeper.GetBalance(testutil.NewContext(app, 1), recipient, "naet")
	require.Equal(t, before, after)
}

func TestWrongChainIDSignedTxFailsBeforeBalanceMutation(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-wrong-chain")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	beforeSender := app.BankKeeper.GetBalance(ctx, sender, "naet")
	beforeRecipient := app.BankKeeper.GetBalance(ctx, recipient, "naet")
	_, sequenceBefore := testutil.AccountNumberAndSequence(t, app, ctx, sender)

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	txBytes := testutil.EncodeSignedTxWithChainID(
		t,
		app,
		ctx,
		senderPriv,
		[]sdk.Msg{msg},
		sdk.NewCoins(sdk.NewInt64Coin("naet", 10)),
		200_000,
		"wrong-chain-id",
	)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "tx signed for a different chain-id must fail")

	afterCtx := testutil.NewContext(app, 1)
	require.Equal(t, beforeSender, app.BankKeeper.GetBalance(afterCtx, sender, "naet"))
	require.Equal(t, beforeRecipient, app.BankKeeper.GetBalance(afterCtx, recipient, "naet"))
	_, sequenceAfter := testutil.AccountNumberAndSequence(t, app, afterCtx, sender)
	require.Equal(t, sequenceBefore, sequenceAfter)
}

func TestMissingAndInvalidFeeTxsFailBeforeBalanceMutation(t *testing.T) {
	tests := []struct {
		name  string
		fee   sdk.Coins
		chain string
	}{
		{name: "missing fee", fee: sdk.Coins{}, chain: "aetheris-integration-missing-fee"},
		{name: "invalid fee denom", fee: sdk.NewCoins(sdk.NewInt64Coin("uatom", 10)), chain: "aetheris-integration-invalid-fee"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := testutil.NewInitializedApp(t, tc.chain)
			ctx := testutil.NewContext(app, 1)
			senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
			_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
			beforeSender := app.BankKeeper.GetBalance(ctx, sender, "naet")
			beforeRecipient := app.BankKeeper.GetBalance(ctx, recipient, "naet")
			beforeFees, err := app.FeesKeeper.GetProtocolFeeState(ctx)
			require.NoError(t, err)
			_, sequenceBefore := testutil.AccountNumberAndSequence(t, app, ctx, sender)

			msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
			txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, tc.fee, 200_000)
			res := testutil.FinalizeBlock(t, app, 1, txBytes)
			require.Len(t, res.TxResults, 1)
			require.NotZero(t, res.TxResults[0].Code)

			afterCtx := testutil.NewContext(app, 1)
			require.Equal(t, beforeSender, app.BankKeeper.GetBalance(afterCtx, sender, "naet"))
			require.Equal(t, beforeRecipient, app.BankKeeper.GetBalance(afterCtx, recipient, "naet"))
			afterFees, err := app.FeesKeeper.GetProtocolFeeState(afterCtx)
			require.NoError(t, err)
			require.Equal(t, beforeFees, afterFees)
			_, sequenceAfter := testutil.AccountNumberAndSequence(t, app, afterCtx, sender)
			require.Equal(t, sequenceBefore, sequenceAfter)
		})
	}
}

func TestUserCreatedTokenCannotPayProtocolFeesEvenWhenOwned(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-user-token-fee")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	tfMsgServer := tfkeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createDenom, err := tfMsgServer.CreateDenom(ctx, &tftypes.MsgCreateDenom{
		Creator:  sender.String(),
		Subdenom: "feeasset",
	})
	require.NoError(t, err)
	denom := createDenom.NewTokenDenom
	_, err = tfMsgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        sender.String(),
		Amount:        sdk.NewInt64Coin(denom, 500),
		MintToAddress: sender.String(),
	})
	require.NoError(t, err)

	beforeSenderNative := app.BankKeeper.GetBalance(ctx, sender, "naet")
	beforeSenderFactory := app.BankKeeper.GetBalance(ctx, sender, denom)
	beforeRecipientNative := app.BankKeeper.GetBalance(ctx, recipient, "naet")
	beforeFees, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)
	_, sequenceBefore := testutil.AccountNumberAndSequence(t, app, ctx, sender)

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "owned user-created token must not satisfy protocol fees")

	afterCtx := testutil.NewContext(app, 1)
	require.Equal(t, beforeSenderNative, app.BankKeeper.GetBalance(afterCtx, sender, "naet"))
	require.Equal(t, beforeSenderFactory, app.BankKeeper.GetBalance(afterCtx, sender, denom))
	require.Equal(t, beforeRecipientNative, app.BankKeeper.GetBalance(afterCtx, recipient, "naet"))
	afterFees, err := app.FeesKeeper.GetProtocolFeeState(afterCtx)
	require.NoError(t, err)
	require.Equal(t, beforeFees, afterFees)
	_, sequenceAfter := testutil.AccountNumberAndSequence(t, app, afterCtx, sender)
	require.Equal(t, sequenceBefore, sequenceAfter)
}

func TestInsufficientFeeFundsFailBeforeStateTransition(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-insufficient-fee-funds")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	beforeSender := app.BankKeeper.GetBalance(ctx, sender, "naet")
	beforeRecipient := app.BankKeeper.GetBalance(ctx, recipient, "naet")
	beforeFees, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)
	_, sequenceBefore := testutil.AccountNumberAndSequence(t, app, ctx, sender)

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 1)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "fee payer without enough naet for fee must fail")

	afterCtx := testutil.NewContext(app, 1)
	require.Equal(t, beforeSender, app.BankKeeper.GetBalance(afterCtx, sender, "naet"))
	require.Equal(t, beforeRecipient, app.BankKeeper.GetBalance(afterCtx, recipient, "naet"))
	afterFees, err := app.FeesKeeper.GetProtocolFeeState(afterCtx)
	require.NoError(t, err)
	require.Equal(t, beforeFees, afterFees)
	_, sequenceAfter := testutil.AccountNumberAndSequence(t, app, afterCtx, sender)
	require.Equal(t, sequenceBefore, sequenceAfter)
}

func TestMalformedProtobufTxBytesFailWithoutFeeAccounting(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-malformed-protobuf")
	ctx := testutil.NewContext(app, 1)
	beforeFees, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)

	res := testutil.FinalizeBlock(t, app, 1, []byte{0xff}, []byte{0x0a, 0x80, 0x80}, []byte("not-a-protobuf-tx"))
	require.Len(t, res.TxResults, 3)
	for _, txResult := range res.TxResults {
		require.NotZero(t, txResult.Code)
	}

	afterFees, err := app.FeesKeeper.GetProtocolFeeState(testutil.NewContext(app, 1))
	require.NoError(t, err)
	require.Equal(t, beforeFees, afterFees)
}

func TestNativeFeeDeductionUpdatesCollectorAndAccounting(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-fee-deduction")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	feeCollector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	require.NotNil(t, feeCollector)

	beforeSender := app.BankKeeper.GetBalance(ctx, sender, "naet")
	beforeRecipient := app.BankKeeper.GetBalance(ctx, recipient, "naet")
	beforeCollector := app.BankKeeper.GetBalance(ctx, feeCollector, "naet")
	beforeFees, err := app.FeesKeeper.GetProtocolFeeState(ctx)
	require.NoError(t, err)

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.Zero(t, res.TxResults[0].Code, res.TxResults[0].Log)
	testutil.Commit(t, app)

	afterCtx := testutil.NewContext(app, 2)
	require.Equal(t, beforeSender.Sub(sdk.NewInt64Coin("naet", 200)), app.BankKeeper.GetBalance(afterCtx, sender, "naet"))
	require.Equal(t, beforeRecipient.Add(sdk.NewInt64Coin("naet", 100)), app.BankKeeper.GetBalance(afterCtx, recipient, "naet"))
	collectorAfter := app.BankKeeper.GetBalance(afterCtx, feeCollector, "naet")
	require.True(t, collectorAfter.Amount.GTE(beforeCollector.Amount.AddRaw(100)), "fee collector balance must increase by at least the tx fee")

	state, err := app.FeesKeeper.GetProtocolFeeState(afterCtx)
	require.NoError(t, err)
	require.Equal(t, beforeFees.TotalCollected.Add(sdk.NewInt64Coin("naet", 100)), state.TotalCollected)
	require.Equal(t, beforeFees.ValidatorRewards.Add(sdk.NewInt64Coin("naet", 98)), state.ValidatorRewards)
	require.Equal(t, beforeFees.CommunityPool.Add(sdk.NewInt64Coin("naet", 2)), state.CommunityPool)
	require.NoError(t, state.Validate())

	moduleBalances, err := app.FeesKeeper.ModuleBalances(afterCtx, &feestypes.QueryModuleBalancesRequest{})
	require.NoError(t, err)
	feeCollectorFound := false
	for _, balance := range moduleBalances.Balances {
		if balance.ModuleName != feestypes.FeeCollectorModuleName {
			continue
		}
		feeCollectorFound = true
		require.Equal(t, aetherisaddress.FormatAccAddress(feeCollector), balance.Address)
		require.Equal(t, app.BankKeeper.GetAllBalances(afterCtx, feeCollector), balance.Balance)
	}
	require.True(t, feeCollectorFound, "fee collector module balance must be exposed")
}

func TestBankTransferEmitsDeterministicTransferEvent(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-bank-events")
	ctx := testutil.NewContext(app, 1)
	senderPriv, sender := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))
	_, recipient := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(1_000_000))

	msg := banktypes.NewMsgSend(sender, recipient, sdk.NewCoins(sdk.NewInt64Coin("naet", 123)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, senderPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), 200_000)
	res := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, res.TxResults, 1)
	require.Zero(t, res.TxResults[0].Code, res.TxResults[0].Log)

	requireABCIEvent(t, res.TxResults[0].Events, banktypes.EventTypeTransfer, map[string]string{
		banktypes.AttributeKeySender:    aetherisaddress.FormatAccAddress(sender),
		banktypes.AttributeKeyRecipient: aetherisaddress.FormatAccAddress(recipient),
		sdk.AttributeKeyAmount:          "123naet",
	})
}

func TestTokenfactoryDexFeesCrossModuleLifecycle(t *testing.T) {
	app := testutil.NewInitializedApp(t, "aetheris-integration-3")
	ctx := testutil.NewContext(app, 1)
	adminPriv, admin := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(2_000_000))
	_, trader := testutil.AddFundedSigner(t, app, ctx, sdkmath.NewInt(2_000_000))
	tfMsgServer := tfkeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	dexMsgServer := dexkeeper.NewMsgServerImpl(app.DexKeeper)

	createDenom, err := tfMsgServer.CreateDenom(ctx, &tftypes.MsgCreateDenom{Creator: admin.String(), Subdenom: "silver"})
	require.NoError(t, err)
	denom := createDenom.NewTokenDenom
	_, err = tfMsgServer.Mint(ctx, &tftypes.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 2_000),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)

	createPool, err := dexMsgServer.CreatePool(ctx, &dextypes.MsgCreatePool{
		Creator: admin.String(),
		TokenA:  sdk.NewInt64Coin("naet", 1_000),
		TokenB:  sdk.NewInt64Coin(denom, 1_000),
	})
	require.NoError(t, err)
	_, err = dexMsgServer.SwapExactAmountIn(ctx, &dextypes.MsgSwapExactAmountIn{
		Trader:        admin.String(),
		PoolId:        createPool.PoolId,
		TokenIn:       sdk.NewInt64Coin("naet", 10),
		TokenOutDenom: denom,
		MinAmountOut:  "1",
	})
	require.NoError(t, err)

	denomQuery, err := app.TokenFactoryKeeper.Denom(ctx, &tftypes.QueryDenomRequest{Denom: denom})
	require.NoError(t, err)
	require.Equal(t, aetherisaddress.FormatAccAddress(admin), denomQuery.Metadata.Admin)
	poolQuery, err := app.DexKeeper.Pool(ctx, &dextypes.QueryPoolRequest{PoolId: createPool.PoolId})
	require.NoError(t, err)
	testutil.AssertPoolAccounting(t, app, ctx, poolQuery.Pool)

	msg := banktypes.NewMsgSend(admin, trader, sdk.NewCoins(sdk.NewInt64Coin("naet", 100)))
	txBytes := testutil.EncodeSignedTx(t, app, ctx, adminPriv, []sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), 200_000)
	block := testutil.FinalizeBlock(t, app, 1, txBytes)
	require.Len(t, block.TxResults, 1)
	require.Zero(t, block.TxResults[0].Code, block.TxResults[0].Log)
	testutil.Commit(t, app)

	accounting, err := app.FeesKeeper.Accounting(testutil.NewContext(app, 2), &feestypes.QueryAccountingRequest{})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("naet", 10)), accounting.ProtocolFeeState.TotalCollected)
	require.NoError(t, accounting.ProtocolFeeState.Validate())
}

func requireABCIEvent(t *testing.T, events []abci.Event, eventType string, attrs map[string]string) {
	t.Helper()
	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		found := make(map[string]string, len(event.Attributes))
		for _, attr := range event.Attributes {
			found[attr.Key] = attr.Value
		}
		matches := true
		for key, expected := range attrs {
			if found[key] != expected {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		return
	}
	require.Failf(t, "missing event", "event type %s with attrs %v not emitted in events %+v", eventType, attrs, events)
}
