package services

import (
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func AutoCLIOptions(modules map[string]any) autocli.AppOptions {
	appModules := make(map[string]appmodule.AppModule)
	for _, m := range modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				appModules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               appModules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(modules),
		AddressCodec:          aetraaddress.Codec{},
		ValidatorAddressCodec: aetraaddress.Codec{},
		ConsensusAddressCodec: aetraaddress.Codec{},
	}
}

func RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig, basicManager module.BasicManager) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	basicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}
