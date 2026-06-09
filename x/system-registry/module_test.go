package systemregistry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/x/system-registry/keeper"
	"github.com/sovereign-l1/l1/x/system-registry/types"
)

func TestAppModuleRegistersRuntimeServicesAndCommands(t *testing.T) {
	k := keeper.NewKeeper()
	msgRouter, queryRouter := registerSystemRegistryServices(t, &k)

	require.NotNil(t, msgRouter.Handler(&types.MsgRegisterSystemEntity{}))
	require.NotNil(t, msgRouter.Handler(&types.MsgPauseSystemEntity{}))
	require.NotNil(t, queryRouter.Route("/l1.systemregistry.v1.Query/SystemEntities"))
	require.NotNil(t, queryRouter.Route("/l1.systemregistry.v1.Query/ReservedSystemAddresses"))
	require.NotNil(t, NewAppModule(&k).GetTxCmd())
	require.NotNil(t, NewAppModule(&k).GetQueryCmd())
}

func registerSystemRegistryServices(t *testing.T, k *keeper.Keeper) (*baseapp.MsgServiceRouter, *baseapp.GRPCQueryRouter) {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	NewAppModule(k).RegisterInterfaces(registry)
	appCodec := codec.NewProtoCodec(registry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(registry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	NewAppModule(k).RegisterServices(sdkmodule.NewConfigurator(appCodec, msgRouter, queryRouter))
	return msgRouter, queryRouter
}
