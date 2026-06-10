package keeperwiring

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wiring/storekeys"
)

func TestNewPersistentKeepersUsesCompleteStoreKeySet(t *testing.T) {
	keys := storekeys.NewKVStoreKeys()

	require.NotPanics(t, func() {
		_ = NewPersistentKeepers(keys, nil)
	})
}
