package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestUpdateParamsRequiresAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "orb1notgov",
		Params:    types.DefaultParams(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
}

func TestUpdateParamsAcceptsGovernanceAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.FeesKeeper.Authority(),
		Params:    types.DefaultParams(),
	})
	require.NoError(t, err)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
}

func TestUpdateParamsRejectsInvalidParamsWithoutMutatingState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := feeskeeper.NewMsgServerImpl(app.FeesKeeper)

	invalid := types.DefaultParams()
	invalid.AllowedFeeDenoms = []string{"testtoken"}

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: app.FeesKeeper.Authority(),
		Params:    invalid,
	})
	require.ErrorIs(t, err, types.ErrInvalidParams)

	params, getErr := app.FeesKeeper.GetParams(ctx)
	require.NoError(t, getErr)
	require.Equal(t, types.DefaultParams(), params)
}
