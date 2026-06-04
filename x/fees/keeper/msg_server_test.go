package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func requireEvent(t *testing.T, ctx sdk.Context, eventType string, attrs map[string]string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		if event.Type != eventType {
			continue
		}
		for key, expected := range attrs {
			attr, found := event.GetAttribute(key)
			require.Truef(t, found, "event %s missing attribute %s", eventType, key)
			require.Equal(t, expected, attr.Value)
		}
		return
	}
	require.Failf(t, "missing event", "event type %s not emitted", eventType)
}

func requireNoEvent(t *testing.T, ctx sdk.Context, eventType string) {
	t.Helper()
	for _, event := range ctx.EventManager().Events() {
		require.NotEqual(t, eventType, event.Type)
	}
}

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
	requireNoEvent(t, ctx, types.EventTypeUpdateParams)
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
	requireEvent(t, ctx, types.EventTypeUpdateParams, map[string]string{
		types.AttributeKeyAuthority:             app.FeesKeeper.Authority(),
		types.AttributeKeyAllowedFeeDenom:       types.DefaultParams().AllowedFeeDenoms[0],
		types.AttributeKeyValidatorRewardsRatio: types.DefaultParams().ValidatorRewardsRatio,
		types.AttributeKeyCommunityPoolRatio:    types.DefaultParams().CommunityPoolRatio,
	})
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
	requireNoEvent(t, ctx, types.EventTypeUpdateParams)
}
