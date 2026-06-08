package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStoreCode{}, "l1/contracts/MsgStoreCode", nil)
	cdc.RegisterConcrete(&MsgDeployContract{}, "l1/contracts/MsgDeployContract", nil)
	cdc.RegisterConcrete(&MsgExecuteExternal{}, "l1/contracts/MsgExecuteExternal", nil)
	cdc.RegisterConcrete(&MsgExecuteInternal{}, "l1/contracts/MsgExecuteInternal", nil)
	cdc.RegisterConcrete(&MsgSendInternalMessage{}, "l1/contracts/MsgSendInternalMessage", nil)
	cdc.RegisterConcrete(&MsgUpdateContractParams{}, "l1/contracts/MsgUpdateContractParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgStoreCode{},
		&MsgDeployContract{},
		&MsgExecuteExternal{},
		&MsgExecuteInternal{},
		&MsgSendInternalMessage{},
		&MsgUpdateContractParams{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&StoreCodeResponse{},
		&InstantiateContractResponse{},
		&ExecuteContractResponse{},
		&InternalMessage{},
		&MsgUpdateContractParamsResponse{},
	)
}
