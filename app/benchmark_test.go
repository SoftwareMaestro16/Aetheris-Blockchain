package app

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client/flags"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func BenchmarkEmptyBlockFinalizeCommit(b *testing.B) {
	_, genesis, valSet := deterministicGenesisWithValidator(b)
	genesisBytes, err := cmtjson.MarshalIndent(genesis, "", " ")
	require.NoError(b, err)

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	app := NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, appOptions)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   genesisBytes,
	})
	require.NoError(b, err)

	nextValidatorsHash := valSet.Hash()
	b.ResetTimer()
	for height := int64(1); height <= int64(b.N); height++ {
		_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:             height,
			Hash:               app.LastCommitID().Hash,
			NextValidatorsHash: nextValidatorsHash,
		})
		require.NoError(b, err)
		_, err = app.Commit()
		require.NoError(b, err)
	}
}
