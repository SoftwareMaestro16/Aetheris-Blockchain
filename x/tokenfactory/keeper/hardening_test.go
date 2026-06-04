package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	tokenfactorykeeper "github.com/sovereign-l1/l1/x/tokenfactory/keeper"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func setupTokenFactory(t *testing.T, addrCount int) (*l1app.L1App, sdk.Context, types.MsgServer, []sdk.AccAddress) {
	t.Helper()

	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsIncremental(app, ctx, addrCount, sdkmath.NewInt(1_000_000))
	msgServer := tokenfactorykeeper.NewMsgServerImpl(app.TokenFactoryKeeper)
	return app, ctx, msgServer, addrs
}

func mustCreateFactoryDenom(t *testing.T, msgServer types.MsgServer, ctx sdk.Context, admin sdk.AccAddress, subdenom string) string {
	t.Helper()

	res, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: subdenom,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.NewTokenDenom)
	return res.NewTokenDenom
}

func TestCreateDenomRejectsMalformedAndReservedSubdenoms(t *testing.T) {
	_, ctx, msgServer, addrs := setupTokenFactory(t, 1)
	admin := addrs[0]

	for _, subdenom := range []string{
		"",
		"ab",
		" bad",
		"gold/silver",
		"gold:silver",
		"norb",
		"ORB",
		"lp",
		"lp-1",
	} {
		t.Run(subdenom, func(t *testing.T) {
			_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
				Creator:  admin.String(),
				Subdenom: subdenom,
			})
			require.Error(t, err)
		})
	}

	_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  "not-an-address",
		Subdenom: "gold",
	})
	require.Error(t, err)
}

func TestCreateDenomRejectsDuplicateAndMetadataCollision(t *testing.T) {
	app, ctx, msgServer, addrs := setupTokenFactory(t, 1)
	admin := addrs[0]

	denom := mustCreateFactoryDenom(t, msgServer, ctx, admin, "gold")

	metadata, found := app.BankKeeper.GetDenomMetaData(ctx, denom)
	require.True(t, found)
	require.Equal(t, denom, metadata.Base)
	require.Equal(t, denom, metadata.Display)
	require.Equal(t, denom, metadata.Name)
	require.Equal(t, denom, metadata.Symbol)
	require.NoError(t, metadata.Validate())

	_, err := msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "gold",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom already exists")

	collisionDenom := "factory/" + admin.String() + "/silver"
	app.BankKeeper.SetDenomMetaData(ctx, tokenfactorykeeper.BankMetadata(collisionDenom))
	_, err = msgServer.CreateDenom(ctx, &types.MsgCreateDenom{
		Creator:  admin.String(),
		Subdenom: "silver",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom already exists")
}

func TestMintAndBurnSupplyAccountingAndAuthorization(t *testing.T) {
	app, ctx, msgServer, addrs := setupTokenFactory(t, 2)
	admin, holder := addrs[0], addrs[1]
	denom := mustCreateFactoryDenom(t, msgServer, ctx, admin, "gold")

	_, err := msgServer.Mint(ctx, &types.MsgMint{
		Sender:        holder.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: holder.String(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "only denom admin can mint")

	initialSupply := app.BankKeeper.GetSupply(ctx, denom)
	require.True(t, initialSupply.IsZero())

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 100),
		MintToAddress: holder.String(),
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(denom, 100), app.BankKeeper.GetSupply(ctx, denom))
	require.Equal(t, sdk.NewInt64Coin(denom, 100), app.BankKeeper.GetBalance(ctx, holder, denom))

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin("factory/"+admin.String()+"/missing", 1),
		MintToAddress: admin.String(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "denom not found")

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 70),
		MintToAddress: admin.String(),
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(denom, 170), app.BankKeeper.GetSupply(ctx, denom))

	_, err = msgServer.Burn(ctx, &types.MsgBurn{
		Sender:          admin.String(),
		Amount:          sdk.NewInt64Coin(denom, 20),
		BurnFromAddress: admin.String(),
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(denom, 150), app.BankKeeper.GetSupply(ctx, denom))
	require.Equal(t, sdk.NewInt64Coin(denom, 50), app.BankKeeper.GetBalance(ctx, admin, denom))

	_, err = msgServer.Burn(ctx, &types.MsgBurn{
		Sender:          admin.String(),
		Amount:          sdk.NewInt64Coin(denom, 1_000),
		BurnFromAddress: admin.String(),
	})
	require.Error(t, err)
	require.Equal(t, sdk.NewInt64Coin(denom, 150), app.BankKeeper.GetSupply(ctx, denom))
	require.Equal(t, sdk.NewInt64Coin(denom, 50), app.BankKeeper.GetBalance(ctx, admin, denom))
}

func TestChangeAdminLifecycle(t *testing.T) {
	app, ctx, msgServer, addrs := setupTokenFactory(t, 2)
	admin, nextAdmin := addrs[0], addrs[1]
	denom := mustCreateFactoryDenom(t, msgServer, ctx, admin, "gold")

	_, err := msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender:   admin.String(),
		Denom:    denom,
		NewAdmin: nextAdmin.String(),
	})
	require.NoError(t, err)

	queryRes, err := app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: denom})
	require.NoError(t, err)
	require.Equal(t, nextAdmin.String(), queryRes.Metadata.Admin)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        admin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: admin.String(),
	})
	require.Error(t, err)

	_, err = msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender:   nextAdmin.String(),
		Denom:    denom,
		NewAdmin: "not-an-address",
	})
	require.Error(t, err)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        nextAdmin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: nextAdmin.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender: nextAdmin.String(),
		Denom:  denom,
	})
	require.NoError(t, err)

	meta, found, err := app.TokenFactoryKeeper.GetDenom(ctx, denom)
	require.NoError(t, err)
	require.True(t, found)
	require.Empty(t, meta.Admin)

	_, err = msgServer.Mint(ctx, &types.MsgMint{
		Sender:        nextAdmin.String(),
		Amount:        sdk.NewInt64Coin(denom, 1),
		MintToAddress: nextAdmin.String(),
	})
	require.Error(t, err)

	_, err = msgServer.ChangeAdmin(ctx, &types.MsgChangeAdmin{
		Sender:   nextAdmin.String(),
		Denom:    denom,
		NewAdmin: admin.String(),
	})
	require.Error(t, err)
}

func TestQueryDenomsPaginationBoundsResponse(t *testing.T) {
	app, ctx, msgServer, addrs := setupTokenFactory(t, 1)
	admin := addrs[0]

	for i := range 105 {
		mustCreateFactoryDenom(t, msgServer, ctx, admin, fmt.Sprintf("denom%03d", i))
	}

	res, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{})
	require.NoError(t, err)
	require.Len(t, res.Denoms, 100)
	require.NotNil(t, res.Pagination)
	require.NotEmpty(t, res.Pagination.NextKey)

	nextRes, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &querytypes.PageRequest{Key: res.Pagination.NextKey, Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, nextRes.Denoms, 5)
	require.Empty(t, nextRes.Pagination.NextKey)

	offsetRes, err := app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &querytypes.PageRequest{Offset: 100, Limit: 10, CountTotal: true},
	})
	require.NoError(t, err)
	require.Len(t, offsetRes.Denoms, 5)
	require.Equal(t, uint64(105), offsetRes.Pagination.Total)

	_, err = app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &querytypes.PageRequest{Key: res.Pagination.NextKey, Offset: 1, Limit: 10},
	})
	require.Error(t, err)

	_, err = app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &querytypes.PageRequest{Key: []byte{0x02}, Limit: 10},
	})
	require.Error(t, err)

	_, err = app.TokenFactoryKeeper.Denoms(ctx, &types.QueryDenomsRequest{
		Pagination: &querytypes.PageRequest{Reverse: true, Limit: 10},
	})
	require.Error(t, err)
}

func TestQueryDenomRejectsMalformedDenom(t *testing.T) {
	app, ctx, _, _ := setupTokenFactory(t, 1)

	_, err := app.TokenFactoryKeeper.Denom(ctx, &types.QueryDenomRequest{Denom: "factory/not-an-address/gold"})
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "invalid")
}

func TestGenesisImportExportAllowsRenouncedAndTransferredAdmin(t *testing.T) {
	app, ctx, _, addrs := setupTokenFactory(t, 2)
	creator, transferredAdmin := addrs[0], addrs[1]
	gs := types.GenesisState{Denoms: []types.DenomAuthorityMetadata{
		{Denom: "factory/" + creator.String() + "/alpha", Admin: creator.String()},
		{Denom: "factory/" + creator.String() + "/beta", Admin: transferredAdmin.String()},
		{Denom: "factory/" + creator.String() + "/gamma", Admin: ""},
	}}

	require.NoError(t, gs.Validate())
	require.NotPanics(t, func() {
		app.TokenFactoryKeeper.InitGenesis(ctx, gs)
	})

	exported := app.TokenFactoryKeeper.ExportGenesis(ctx)
	require.ElementsMatch(t, gs.Denoms, exported.Denoms)
}
