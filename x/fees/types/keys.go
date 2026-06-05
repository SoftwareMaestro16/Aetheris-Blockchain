package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName = "fees"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	ParamsKey           = []byte{0x01}
	ProtocolFeeStateKey = []byte{0x02}
	BlockTxCountKey     = []byte{0x03}
	SenderTxCountPrefix = []byte{0x04}
)

const (
	BondDenom              = appparams.BaseDenom
	MaxAllowedFeeDenomsV1  = 1
	MaxMinFeeAmountV1      = "1000000000000000000"
	MaxFeeAmountV1         = "1000000000000000000"
	PrototypeBaseFeeAmount = "1000000"
	PrototypeBaseFeeCoin   = PrototypeBaseFeeAmount + BondDenom
	PrototypeMinGasPriceV1 = "0" + BondDenom
)
