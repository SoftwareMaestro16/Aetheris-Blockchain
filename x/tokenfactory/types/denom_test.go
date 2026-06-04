package types

import (
	"bytes"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateSubdenomHardening(t *testing.T) {
	valid := []string{
		"gold",
		"silver-token_1",
		"Alpha.2",
	}
	for _, subdenom := range valid {
		t.Run("valid "+subdenom, func(t *testing.T) {
			require.NoError(t, ValidateSubdenom(subdenom))
		})
	}

	invalid := []string{
		"",
		"ab",
		"_bad",
		" bad",
		"gold/silver",
		"gold:silver",
		"norb",
		"ORB",
		"lp",
		"lp-1",
		strings.Repeat("a", MaxSubdenomLength+1),
	}
	for _, subdenom := range invalid {
		t.Run("invalid "+subdenom, func(t *testing.T) {
			require.Error(t, ValidateSubdenom(subdenom))
		})
	}
}

func TestFactoryDenomValidation(t *testing.T) {
	creator := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()

	denom, err := BuildFactoryDenom(creator, "gold")
	require.NoError(t, err)
	require.Equal(t, "factory/"+creator+"/gold", denom)

	parsedCreator, subdenom, err := ParseFactoryDenom(denom)
	require.NoError(t, err)
	require.Equal(t, creator, parsedCreator.String())
	require.Equal(t, "gold", subdenom)
	require.NoError(t, ValidateFactoryDenom(denom))

	invalid := []string{
		"norb",
		"lp/1",
		"factory/not-an-address/gold",
		"factory/" + creator + "/gold/extra",
		"factory/" + creator + "/norb",
		"factory/" + creator + "/lp-1",
	}
	for _, denom := range invalid {
		t.Run(denom, func(t *testing.T) {
			require.Error(t, ValidateFactoryDenom(denom))
		})
	}
}
