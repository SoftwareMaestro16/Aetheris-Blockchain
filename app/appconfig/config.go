package appconfig

import (
	clienthelpers "cosmossdk.io/client/v2/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const AppName = appparams.ChainName

const (
	// SDKBech32AccountPrefix is a Cosmos SDK compatibility prefix only.
	// User-facing Aetra addresses use app/addressing's AE... base64url format.
	SDKBech32AccountPrefix   = "ae"
	SDKBech32ValidatorPrefix = "aevaloper"
	SDKBech32ConsensusPrefix = "aevalcons"
	BondDenom                = appparams.BaseDenom
)

const (
	// Deprecated: use SDKBech32AccountPrefix for SDK compatibility code, or
	// app/addressing FormatUserFriendly/Parse for user-facing addresses.
	AccountAddressPrefix = SDKBech32AccountPrefix
	// Deprecated: use SDKBech32ValidatorPrefix.
	ValidatorAddressPrefix = SDKBech32ValidatorPrefix
	// Deprecated: use SDKBech32ConsensusPrefix.
	ConsensusAddressPrefix = SDKBech32ConsensusPrefix
)

func ConfigureSDK(homeName string) string {
	nodeHome, err := clienthelpers.GetNodeHomeDirectory(homeName)
	if err != nil {
		panic(err)
	}
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(SDKBech32AccountPrefix, SDKBech32AccountPrefix+"pub")
	cfg.SetBech32PrefixForValidator(SDKBech32ValidatorPrefix, SDKBech32ValidatorPrefix+"pub")
	cfg.SetBech32PrefixForConsensusNode(SDKBech32ConsensusPrefix, SDKBech32ConsensusPrefix+"pub")
	sdk.DefaultBondDenom = BondDenom
	return nodeHome
}
