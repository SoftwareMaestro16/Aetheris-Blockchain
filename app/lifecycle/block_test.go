package lifecycle

import (
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
)

func TestFinalizeBlockCallsFinalizerAndReturnsError(t *testing.T) {
	expected := errors.New("finalize failed")
	called := false

	res, err := FinalizeBlock(&abci.RequestFinalizeBlock{Height: 10}, func(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
		called = true
		require.Equal(t, int64(10), req.Height)
		return nil, expected
	})

	require.True(t, called)
	require.Nil(t, res)
	require.ErrorIs(t, err, expected)
}
