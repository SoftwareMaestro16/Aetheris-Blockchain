package modulewiring

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"
)

func TestNewBasicManagerAcceptsEmptyManager(t *testing.T) {
	manager := module.NewManager()

	require.NotPanics(t, func() {
		_ = NewBasicManager(manager, codec.NewLegacyAmino(), codectypes.NewInterfaceRegistry())
	})
}
