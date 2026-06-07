package app

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/app/lifecycle"
)

func (app *L1App) Name() string { return app.BaseApp.Name() }

func (app *L1App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

func (app *L1App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

func (app *L1App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *L1App) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return lifecycle.FinalizeBlock(req, app.BaseApp.FinalizeBlock)
}

func (a *L1App) Configurator() module.Configurator {
	return a.configurator
}

func (app *L1App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return lifecycle.InitChain(ctx, req, lifecycle.InitChainDependencies{
		AppCodec:      app.appCodec,
		ModuleManager: app.ModuleManager,
		SetModuleVersionMap: func(ctx sdk.Context, versionMap module.VersionMap) error {
			return app.UpgradeKeeper.SetModuleVersionMap(ctx, versionMap)
		},
		ValidateGenesis:              app.validateAetraGenesis,
		EnsureCoreGenesisCollections: app.ensureCoreGenesisCollections,
	})
}
