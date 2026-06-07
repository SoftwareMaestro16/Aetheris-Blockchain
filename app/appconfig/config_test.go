package appconfig

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestConfigureSDKSetsSDKBech32CompatibilityAndBondDenom(t *testing.T) {
	home := ConfigureSDK(".aetra")

	require.True(t, strings.HasSuffix(home, ".aetra"), home)
	require.Equal(t, SDKBech32AccountPrefix, sdk.GetConfig().GetBech32AccountAddrPrefix())
	require.Equal(t, SDKBech32AccountPrefix, sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	require.Equal(t, SDKBech32AccountPrefix, sdk.GetConfig().GetBech32ConsensusAddrPrefix())
	require.Equal(t, appparams.BaseDenom, sdk.DefaultBondDenom)
}

func TestUserFacingAddressFormatIsAEBase64URLNotSDKBech32(t *testing.T) {
	addr := sdk.AccAddress([]byte{
		0x01, 0x02, 0x03, 0x04, 0x05,
		0x06, 0x07, 0x08, 0x09, 0x0a,
		0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14,
	})

	userFacing, err := addressing.FormatUserFriendly(addr)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(userFacing, "AE"))
	require.Regexp(t, `^[A-Za-z0-9_-]{48}$`, userFacing)
	require.False(t, strings.HasPrefix(userFacing, SDKBech32AccountPrefix+"1"))
	require.Equal(t, "AE", AccountAddressPrefix)
	require.Equal(t, "AE", ValidatorAddressPrefix)
	require.Equal(t, "AE", ConsensusAddressPrefix)
}
