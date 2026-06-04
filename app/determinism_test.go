package app

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisJSONIsDeterministic(t *testing.T) {
	_, firstGenesis := setup(true, 5)
	firstBytes, err := json.Marshal(firstGenesis)
	require.NoError(t, err)
	firstHash := sha256.Sum256(firstBytes)

	for i := 0; i < 5; i++ {
		app, genesis := setup(true, 5)
		genesisBytes, err := json.Marshal(genesis)
		require.NoError(t, err)
		require.Equal(t, firstBytes, genesisBytes, "default genesis JSON changed on iteration %d", i)
		require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

		hash := sha256.Sum256(genesisBytes)
		require.Equal(t, firstHash, hash)
	}
}
