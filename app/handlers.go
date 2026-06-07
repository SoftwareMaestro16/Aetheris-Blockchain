package app

import (
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/sovereign-l1/l1/app/txhandlers"
)

func (app *L1App) setAnteHandler(txConfig client.TxConfig) {
	anteHandler := txhandlers.NewAnteHandler(txConfig, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper)
	app.SetAnteHandler(app.FeesKeeper.AnteHandlerDecorator(anteHandler))
}

func (app *L1App) setPostHandler() {
	app.SetPostHandler(txhandlers.NewPostHandler())
}
