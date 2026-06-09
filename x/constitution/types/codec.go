package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgProposeConstitutionAmendment{}, "l1/constitution/MsgProposeConstitutionAmendment", nil)
	cdc.RegisterConcrete(&MsgVoteConstitutionAmendment{}, "l1/constitution/MsgVoteConstitutionAmendment", nil)
	cdc.RegisterConcrete(&MsgExecuteConstitutionAmendment{}, "l1/constitution/MsgExecuteConstitutionAmendment", nil)
	cdc.RegisterConcrete(&MsgCancelConstitutionAmendment{}, "l1/constitution/MsgCancelConstitutionAmendment", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgProposeConstitutionAmendment{},
		&MsgVoteConstitutionAmendment{},
		&MsgExecuteConstitutionAmendment{},
		&MsgCancelConstitutionAmendment{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&MsgProposeConstitutionAmendmentResponse{},
		&MsgVoteConstitutionAmendmentResponse{},
		&MsgExecuteConstitutionAmendmentResponse{},
		&MsgCancelConstitutionAmendmentResponse{},
	)
}
