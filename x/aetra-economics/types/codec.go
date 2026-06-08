package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateEconomicsParams{}, "l1/aetraeconomics/MsgUpdateEconomicsParams", nil)
	cdc.RegisterConcrete(&MsgApplyEpochEconomics{}, "l1/aetraeconomics/MsgApplyEpochEconomics", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateEconomicsParams{}, &MsgApplyEpochEconomics{})
	registry.RegisterImplementations((*txtypes.MsgResponse)(nil), &MsgUpdateEconomicsParamsResponse{}, &MsgApplyEpochEconomicsResponse{})
}
