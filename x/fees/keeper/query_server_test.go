package keeper_test

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func TestParamsQueryReturnsCurrentParams(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	res, err := app.FeesKeeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), res.Params)
}

func TestParamsQueryRejectsNilRequest(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)

	_, err := app.FeesKeeper.Params(ctx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
