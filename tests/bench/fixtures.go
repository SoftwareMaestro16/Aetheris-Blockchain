package bench

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	l1app "github.com/sovereign-l1/l1/app"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

const ChainID = "orbitalis-bench-1"

var ObjectSizes = []int{1_000, 10_000, 100_000}

type GenesisFixture struct {
	App          *l1app.L1App
	GenesisBytes []byte
	ValSetHash   []byte
	Sender       sdk.AccAddress
	Recipient    sdk.AccAddress
	SenderPriv   cryptotypes.PrivKey
}

func NewApp(b *testing.B) *l1app.L1App {
	b.Helper()

	return l1app.NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		simtestutil.AppOptionsMap{flags.FlagHome: b.TempDir()},
		bam.SetChainID(ChainID),
	)
}

func InitializedApp(b *testing.B) (*l1app.L1App, sdk.Context) {
	b.Helper()

	fixture := GenesisWithLoad(b, 0, 0)
	app := InitApp(b, fixture.GenesisBytes)
	return app, app.NewContext(false)
}

func GenesisWithLoad(b *testing.B, denoms, pools int) GenesisFixture {
	b.Helper()

	app := NewApp(b)
	senderPriv := secp256k1.GenPrivKeyFromSecret([]byte("orbitalis benchmark sender"))
	recipientPriv := secp256k1.GenPrivKeyFromSecret([]byte("orbitalis benchmark recipient"))
	sender := sdk.AccAddress(senderPriv.PubKey().Address())
	recipient := sdk.AccAddress(recipientPriv.PubKey().Address())

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(b, err)

	genAccs := []authtypes.GenesisAccount{
		authtypes.NewBaseAccount(sender, senderPriv.PubKey(), 0, 0),
		authtypes.NewBaseAccount(recipient, recipientPriv.PubKey(), 1, 0),
	}
	balances := banktypes.Balance{
		Address: sender.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(l1app.BondDenom, sdkmath.NewInt(1_000_000_000_000_000_000))),
	}

	genesis := app.DefaultGenesis()
	genesis, err = simtestutil.GenesisStateWithValSet(app.AppCodec(), genesis, valSet, genAccs, balances)
	require.NoError(b, err)

	if denoms > 0 {
		genesis[tokenfactorytypes.ModuleName] = app.AppCodec().MustMarshalJSON(TokenFactoryGenesis(sender.String(), denoms))
	}
	if pools > 0 {
		genesis[dextypes.ModuleName] = app.AppCodec().MustMarshalJSON(DexGenesis(pools))
	}

	genesisBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(b, err)
	return GenesisFixture{
		App:          app,
		GenesisBytes: genesisBytes,
		ValSetHash:   valSet.Hash(),
		Sender:       sender,
		Recipient:    recipient,
		SenderPriv:   senderPriv,
	}
}

func InitApp(b *testing.B, genesisBytes []byte) *l1app.L1App {
	b.Helper()

	app := NewApp(b)
	_, err := app.InitChain(&abci.RequestInitChain{
		ChainId:         ChainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: ConsensusParams(),
		AppStateBytes:   genesisBytes,
	})
	require.NoError(b, err)
	return app
}

func FinalizeEmptyBlock(b *testing.B, app *l1app.L1App, height int64, valSetHash []byte) []byte {
	b.Helper()

	resp, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             height,
		Hash:               []byte(fmt.Sprintf("bench-block-%d", height)),
		NextValidatorsHash: valSetHash,
	})
	require.NoError(b, err)
	_, err = app.Commit()
	require.NoError(b, err)
	return resp.AppHash
}

func TokenFactoryGenesis(admin string, count int) *tokenfactorytypes.GenesisState {
	denoms := make([]tokenfactorytypes.DenomAuthorityMetadata, count)
	for i := 0; i < count; i++ {
		denoms[i] = tokenfactorytypes.DenomAuthorityMetadata{
			Denom: fmt.Sprintf("factory/%s/bench%06d", admin, i),
			Admin: admin,
		}
	}
	return &tokenfactorytypes.GenesisState{Denoms: denoms}
}

func DexGenesis(count int) *dextypes.GenesisState {
	pools := make([]dextypes.Pool, count)
	for i := 0; i < count; i++ {
		id := uint64(i + 1)
		pools[i] = dextypes.Pool{
			Id:          id,
			Denom0:      fmt.Sprintf("bench%06da", i),
			Denom1:      fmt.Sprintf("bench%06db", i),
			Reserve0:    "1000000000",
			Reserve1:    "2000000000",
			TotalShares: "1000000000",
			LpDenom:     fmt.Sprintf("%s/%d", dextypes.LPDenomPrefix, id),
		}
	}
	return &dextypes.GenesisState{NextPoolId: uint64(count + 1), Pools: pools}
}

func SeedTokenFactoryDenoms(b *testing.B, app *l1app.L1App, ctx context.Context, admin string, count int) {
	b.Helper()

	gs := TokenFactoryGenesis(admin, count)
	app.TokenFactoryKeeper.InitGenesis(ctx, *gs)
}

func SeedDexPools(b *testing.B, app *l1app.L1App, ctx context.Context, count int) {
	b.Helper()

	gs := DexGenesis(count)
	app.DexKeeper.InitGenesis(ctx, *gs)
}

func MsgSendTxs(b *testing.B, app *l1app.L1App, priv cryptotypes.PrivKey, from, to sdk.AccAddress, count int) [][]byte {
	b.Helper()

	txs := make([][]byte, count)
	for i := 0; i < count; i++ {
		txs[i] = MsgSendTx(b, app, priv, from, to, uint64(i))
	}
	return txs
}

func MsgSendTx(b *testing.B, app *l1app.L1App, priv cryptotypes.PrivKey, from, to sdk.AccAddress, sequence uint64) []byte {
	b.Helper()

	txConfig := app.TxConfig()
	txBuilder := txConfig.NewTxBuilder()
	msg := banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewInt64Coin(l1app.BondDenom, 1)))
	require.NoError(b, txBuilder.SetMsgs(msg))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200_000)
	txBuilder.SetMemo("")

	signMode, err := authsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(b, err)
	sig := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signMode,
		},
		Sequence: sequence,
	}
	require.NoError(b, txBuilder.SetSignatures(sig))

	signerData := authsigning.SignerData{
		Address:       from.String(),
		ChainID:       ChainID,
		AccountNumber: 0,
		Sequence:      sequence,
		PubKey:        priv.PubKey(),
	}
	signBytes, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signMode,
		signerData,
		txBuilder.GetTx(),
	)
	require.NoError(b, err)

	signature, err := priv.Sign(signBytes)
	require.NoError(b, err)
	sig.Data.(*signing.SingleSignatureData).Signature = signature
	require.NoError(b, txBuilder.SetSignatures(sig))

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(b, err)
	return txBytes
}

func ConsensusParams() *cmtproto.ConsensusParams {
	params := *simtestutil.DefaultConsensusParams
	block := *params.Block
	block.MaxBytes = 100_000_000
	block.MaxGas = -1
	params.Block = &block
	return &params
}

func ShuffleBytes(in [][]byte, seed int64) [][]byte {
	out := append([][]byte(nil), in...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

func ValidatorSetHash(b *testing.B) []byte {
	b.Helper()

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(b, err)
	return cmttypes.NewValidatorSet(valSet.Validators).Hash()
}
