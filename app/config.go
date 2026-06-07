package app

import (
	"github.com/sovereign-l1/l1/app/appconfig"
)

const appName = appconfig.AppName

const (
	SDKBech32AccountPrefix   = appconfig.SDKBech32AccountPrefix
	SDKBech32ValidatorPrefix = appconfig.SDKBech32ValidatorPrefix
	SDKBech32ConsensusPrefix = appconfig.SDKBech32ConsensusPrefix
	BondDenom                = appconfig.BondDenom
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

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome = appconfig.ConfigureSDK(".aetra")
