package keeper_test

import (
	"errors"
	"testing"

	sdkmath "cosmossdk.io/math"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

type feeTx struct {
	fees sdk.Coins
}

func (tx feeTx) GetMsgs() []sdk.Msg {
	return nil
}

func (tx feeTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func (tx feeTx) GetGas() uint64 {
	return 100_000
}

func (tx feeTx) GetFee() sdk.Coins {
	return tx.fees
}

func (tx feeTx) FeePayer() []byte {
	return nil
}

func (tx feeTx) FeeGranter() []byte {
	return nil
}

type noFeeTx struct{}

func (tx noFeeTx) GetMsgs() []sdk.Msg {
	return nil
}

func (tx noFeeTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func TestAnteHandlerDecoratorFeePolicy(t *testing.T) {
	tests := []struct {
		name         string
		tx           sdk.Tx
		wantErr      string
		wantNextCall bool
	}{
		{
			name:         "accepts native fee denom",
			tx:           feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))},
			wantNextCall: true,
		},
		{
			name:         "accepts empty fee list",
			tx:           feeTx{fees: sdk.Coins{}},
			wantNextCall: true,
		},
		{
			name:         "accepts nil fee list",
			tx:           feeTx{},
			wantNextCall: true,
		},
		{
			name:         "accepts zero native fee coin",
			tx:           feeTx{fees: sdk.Coins{sdk.NewInt64Coin(types.BondDenom, 0)}},
			wantNextCall: true,
		},
		{
			name:    "rejects non native fee denom",
			tx:      feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin("uatom", 1))},
			wantErr: "fee denom uatom not accepted; use norb",
		},
		{
			name:    "rejects mixed native and non native fee denoms",
			tx:      feeTx{fees: sdk.Coins{sdk.NewInt64Coin(types.BondDenom, 1), sdk.NewInt64Coin("testtoken", 1)}},
			wantErr: "fee denom testtoken not accepted; use norb",
		},
		{
			name:    "rejects malformed fee coin",
			tx:      feeTx{fees: sdk.Coins{{Denom: "!", Amount: sdkmath.NewInt(1)}}},
			wantErr: "fee coins must be valid",
		},
		{
			name: "rejects duplicate fee denom entries",
			tx: feeTx{fees: sdk.Coins{
				sdk.NewInt64Coin(types.BondDenom, 1),
				sdk.NewInt64Coin(types.BondDenom, 2),
			}},
			wantErr: "fee coins must be valid",
		},
		{
			name:    "rejects transaction without fee interface",
			tx:      noFeeTx{},
			wantErr: "transaction must expose fees",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := l1app.Setup(t, false)
			ctx := app.NewContext(false)

			called := false
			next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
				called = true
				return ctx, nil
			}

			_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, tc.tx, false)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, types.ErrInvalidFee)
				require.Contains(t, err.Error(), tc.wantErr)
			}
			require.Equal(t, tc.wantNextCall, called)
		})
	}
}

func TestAnteHandlerDecoratorPropagatesNextError(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	nextErr := errors.New("next failed")

	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nextErr
	}

	_, err := app.FeesKeeper.AnteHandlerDecorator(next)(ctx, feeTx{fees: sdk.NewCoins(sdk.NewInt64Coin(types.BondDenom, 1))}, false)
	require.ErrorIs(t, err, nextErr)
}
