package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateEmissionsParams{}, "l1/emissions/MsgUpdateEmissionsParams", nil)
	cdc.RegisterConcrete(&MsgFinalizeEmissionEpoch{}, "l1/emissions/MsgFinalizeEmissionEpoch", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateEmissionsParams{},
		&MsgFinalizeEmissionEpoch{},
	)
}
