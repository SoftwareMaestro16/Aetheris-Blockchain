package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateStakingPolicyParams{}, "l1/aetrastakingpolicy/MsgUpdateStakingPolicyParams", nil)
	cdc.RegisterConcrete(&MsgRegisterValidatorIdentity{}, "l1/aetrastakingpolicy/MsgRegisterValidatorIdentity", nil)
	cdc.RegisterConcrete(&MsgAcknowledgeConcentrationWarning{}, "l1/aetrastakingpolicy/MsgAcknowledgeConcentrationWarning", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateStakingPolicyParams{}, &MsgRegisterValidatorIdentity{}, &MsgAcknowledgeConcentrationWarning{})
	registry.RegisterImplementations((*txtypes.MsgResponse)(nil), &MsgUpdateStakingPolicyParamsResponse{}, &MsgRegisterValidatorIdentityResponse{}, &MsgAcknowledgeConcentrationWarningResponse{})
}
