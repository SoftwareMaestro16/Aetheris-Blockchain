package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterSystemEntity{}, "l1/systemregistry/MsgRegisterSystemEntity", nil)
	cdc.RegisterConcrete(&MsgUpdateSystemEntity{}, "l1/systemregistry/MsgUpdateSystemEntity", nil)
	cdc.RegisterConcrete(&MsgPauseSystemEntity{}, "l1/systemregistry/MsgPauseSystemEntity", nil)
	cdc.RegisterConcrete(&MsgResumeSystemEntity{}, "l1/systemregistry/MsgResumeSystemEntity", nil)
	cdc.RegisterConcrete(&MsgDeprecateSystemEntity{}, "l1/systemregistry/MsgDeprecateSystemEntity", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterSystemEntity{},
		&MsgUpdateSystemEntity{},
		&MsgPauseSystemEntity{},
		&MsgResumeSystemEntity{},
		&MsgDeprecateSystemEntity{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&MsgRegisterSystemEntityResponse{},
		&MsgUpdateSystemEntityResponse{},
		&MsgPauseSystemEntityResponse{},
		&MsgResumeSystemEntityResponse{},
		&MsgDeprecateSystemEntityResponse{},
	)
}
