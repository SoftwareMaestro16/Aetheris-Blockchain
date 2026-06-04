package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func TestDenomsQueryReturnsEmptyState(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	res, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{})
	require.NoError(t, err)
	require.Empty(t, res.Denoms)
}

func TestDenomQueryReturnsMetadata(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)

	createRes, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "querygold",
	})
	require.NoError(t, err)

	res, err := app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: createRes.NewTokenDenom})
	require.NoError(t, err)
	require.Equal(t, createRes.NewTokenDenom, res.Metadata.Denom)
	require.Equal(t, admin.String(), res.Metadata.Admin)
}

func TestDenomQueryErrorsAreGrpcStatusCompatible(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]
	missingDenom := fmt.Sprintf("factory/%s/missing", admin.String())

	_, err := app.TokenFactoryKeeper.Denom(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: "bad denom"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: missingDenom})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestDenomsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.TokenFactoryKeeper.Denoms(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestDenomsQueryEnforcesPrototypeLimit(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	admin := l1app.AddTestAddrsIncremental(app, ctx, 1, sdkmath.NewInt(1_000_000))[0]

	for i := 0; i < types.MaxQueryDenoms+1; i++ {
		denom := fmt.Sprintf("factory/%s/querygold%d", admin.String(), i)
		require.NoError(t, app.TokenFactoryKeeper.SetDenom(ctx, types.DenomAuthorityMetadata{
			Denom: denom,
			Admin: admin.String(),
		}))
	}

	_, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{})
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
}
