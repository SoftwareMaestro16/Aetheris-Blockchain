package contracts

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/contracts/keeper"
)

func TestAppModuleRegistersTxAndQueryServices(t *testing.T) {
	k := keeper.NewKeeper()
	registry := codectypes.NewInterfaceRegistry()
	NewAppModule(&k).RegisterInterfaces(registry)
	appCodec := codec.NewProtoCodec(registry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	queryRouter.SetInterfaceRegistry(registry)
	NewAppModule(&k).RegisterServices(module.NewConfigurator(appCodec, msgRouter, queryRouter))

	for _, route := range []string{
		"/l1.contracts.v1.Query/Params",
		"/l1.contracts.v1.Query/Code",
		"/l1.contracts.v1.Query/Codes",
		"/l1.contracts.v1.Query/Contract",
		"/l1.contracts.v1.Query/Contracts",
		"/l1.contracts.v1.Query/ContractStorage",
		"/l1.contracts.v1.Query/ContractReceipts",
		"/l1.contracts.v1.Query/ContractQueue",
		"/l1.contracts.v1.Query/ContractEvents",
		"/l1.contracts.v1.Query/ContractStateRoot",
	} {
		require.NotNil(t, queryRouter.Route(route), route)
	}
}
