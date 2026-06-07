package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoCLIOptionsUsesAetraAddressCodecs(t *testing.T) {
	opts := AutoCLIOptions(map[string]any{})

	require.NotNil(t, opts.AddressCodec)
	require.NotNil(t, opts.ValidatorAddressCodec)
	require.NotNil(t, opts.ConsensusAddressCodec)
	require.Empty(t, opts.Modules)
}
