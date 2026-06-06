package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName = "feecollector"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	CollectorModuleName  = ModuleName
	TreasuryModuleName   = "feecollector_treasury"
	ProtectionModuleName = "feecollector_protection"
)

var (
	ParamsKey              = []byte{0x01}
	FeeBalancesKey         = []byte{0x02}
	PendingDistributionKey = []byte{0x03}
	FeeHistoryPrefix       = []byte{0x04}
)

const (
	BaseDenom          = appparams.BaseDenom
	BasisPoints uint32 = 10_000

	FeeTypeGas        = "gas"
	FeeTypeForwarding = "forwarding"
	FeeTypeProtocol   = "protocol"
)
