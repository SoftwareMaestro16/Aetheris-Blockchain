package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitConfigChange{}, "l1/config/MsgSubmitConfigChange", nil)
	cdc.RegisterConcrete(&MsgApproveConfigChange{}, "l1/config/MsgApproveConfigChange", nil)
	cdc.RegisterConcrete(&MsgRejectConfigChange{}, "l1/config/MsgRejectConfigChange", nil)
	cdc.RegisterConcrete(&MsgExecuteConfigChange{}, "l1/config/MsgExecuteConfigChange", nil)
	cdc.RegisterConcrete(&MsgCancelConfigChange{}, "l1/config/MsgCancelConfigChange", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitConfigChange{},
		&MsgApproveConfigChange{},
		&MsgRejectConfigChange{},
		&MsgExecuteConfigChange{},
		&MsgCancelConfigChange{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&MsgSubmitConfigChangeResponse{},
		&MsgApproveConfigChangeResponse{},
		&MsgRejectConfigChangeResponse{},
		&MsgExecuteConfigChangeResponse{},
		&MsgCancelConfigChangeResponse{},
	)
}
