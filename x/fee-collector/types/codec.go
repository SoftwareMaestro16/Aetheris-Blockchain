package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDistributeFees{}, "l1/feecollector/MsgDistributeFees", nil)
	cdc.RegisterConcrete(&MsgUpdateFeeDistributionParams{}, "l1/feecollector/MsgUpdateFeeDistributionParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgDistributeFees{},
		&MsgUpdateFeeDistributionParams{},
	)
}
