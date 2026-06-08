package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateValidatorScoreParams{}, "l1/aetravalidatorscore/MsgUpdateValidatorScoreParams", nil)
	cdc.RegisterConcrete(&MsgUpdateValidatorScores{}, "l1/aetravalidatorscore/MsgUpdateValidatorScores", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateValidatorScoreParams{}, &MsgUpdateValidatorScores{})
	registry.RegisterImplementations((*txtypes.MsgResponse)(nil), &MsgUpdateValidatorScoreParamsResponse{}, &MsgUpdateValidatorScoresResponse{})
}
