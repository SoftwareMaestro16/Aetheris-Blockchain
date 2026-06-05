package types

const (
	ModuleName = "dex"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	PoolPrefix        = []byte{0x01}
	NextPoolIDKey     = []byte{0x02}
	PairPrefix        = []byte{0x03}
	ParamsKey         = []byte{0x04}
	DefaultNextPoolID = uint64(1)
)

const (
	LPDenomPrefix        = "lp"
	DefaultSwapFeeBps    = uint32(30)
	DefaultMaxSwapFeeBps = uint32(1_000)
	PoolFeeBps           = int64(DefaultSwapFeeBps)
	BpsDenominator       = int64(10_000)
	DefaultQueryPools    = 50
	MaxQueryPools        = 100
)
