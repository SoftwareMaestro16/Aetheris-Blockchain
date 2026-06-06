package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitTreasurySpend{}, "l1/treasury/MsgSubmitTreasurySpend", nil)
	cdc.RegisterConcrete(&MsgApproveTreasurySpend{}, "l1/treasury/MsgApproveTreasurySpend", nil)
	cdc.RegisterConcrete(&MsgRejectTreasurySpend{}, "l1/treasury/MsgRejectTreasurySpend", nil)
	cdc.RegisterConcrete(&MsgExecuteTreasurySpend{}, "l1/treasury/MsgExecuteTreasurySpend", nil)
	cdc.RegisterConcrete(&MsgCancelTreasurySpend{}, "l1/treasury/MsgCancelTreasurySpend", nil)
	cdc.RegisterConcrete(&MsgUpdateTreasuryParams{}, "l1/treasury/MsgUpdateTreasuryParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitTreasurySpend{},
		&MsgApproveTreasurySpend{},
		&MsgRejectTreasurySpend{},
		&MsgExecuteTreasurySpend{},
		&MsgCancelTreasurySpend{},
		&MsgUpdateTreasuryParams{},
	)
}
