package types

const (
	ModuleName = "stakeconcentration"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	BasisPoints uint32 = 10_000

	AetraValidatorSetPhaseOneMax = 150
	AetraValidatorSetPhaseTwoMax = 250

	AetraPhaseOnePowerCapBps  uint32 = 300
	AetraPhaseTwoPowerCapBps  uint32 = 250
	AetraMatureSetPowerCapBps uint32 = 200
)

var (
	ParamsKey  = []byte{0x01}
	NetworkKey = []byte{0x02}
)
