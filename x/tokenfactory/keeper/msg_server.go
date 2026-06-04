package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) CreateDenom(ctx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidDenom.Wrap("empty create denom request")
	}
	denom, err := m.FullDenom(msg.Creator, msg.Subdenom)
	if err != nil {
		return nil, err
	}
	creator, err := parseAccAddress("creator", msg.Creator)
	if err != nil {
		return nil, err
	}
	if _, found, err := m.GetDenom(ctx, denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrDenomExists.Wrap(denom)
	}
	if _, found := m.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return nil, types.ErrDenomExists.Wrap("bank metadata already exists for " + denom)
	}

	meta := types.DenomAuthorityMetadata{Denom: denom, Admin: creator.String()}
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
	bankMetadata := BankMetadata(denom)
	if err := bankMetadata.Validate(); err != nil {
		return nil, types.ErrInvalidDenom.Wrapf("invalid bank metadata: %v", err)
	}
	m.bankKeeper.SetDenomMetaData(ctx, bankMetadata)
	return &types.MsgCreateDenomResponse{NewTokenDenom: denom}, nil
}

func (m msgServer) Mint(ctx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidDenom.Wrap("empty mint request")
	}
	sender, err := parseAccAddress("sender", msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if err := requireAdmin(meta, sender, "mint"); err != nil {
		return nil, err
	}
	to, err := parseAccAddress("mint_to_address", msg.MintToAddress)
	if err != nil {
		return nil, err
	}
	if err := validateFactoryAmount(meta, msg.Amount, "mint"); err != nil {
		return nil, err
	}

	beforeSupply := m.bankKeeper.GetSupply(ctx, meta.Denom)
	coins := sdk.NewCoins(msg.Amount)
	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, coins); err != nil {
		return nil, err
	}
	afterSupply := m.bankKeeper.GetSupply(ctx, meta.Denom)
	if expected := beforeSupply.Add(msg.Amount); !afterSupply.IsEqual(expected) {
		return nil, types.ErrSupplyInvariant.Wrapf("expected supply %s after mint, got %s", expected, afterSupply)
	}
	return &types.MsgMintResponse{}, nil
}

func (m msgServer) Burn(ctx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidDenom.Wrap("empty burn request")
	}
	sender, err := parseAccAddress("sender", msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if err := requireAdmin(meta, sender, "burn"); err != nil {
		return nil, err
	}
	from, err := parseAccAddress("burn_from_address", msg.BurnFromAddress)
	if err != nil {
		return nil, err
	}
	if !from.Equals(sender) {
		return nil, types.ErrUnauthorized.Wrap("burn_from_address must match sender")
	}
	if err := validateFactoryAmount(meta, msg.Amount, "burn"); err != nil {
		return nil, err
	}

	beforeSupply := m.bankKeeper.GetSupply(ctx, meta.Denom)
	if beforeSupply.IsLT(msg.Amount) {
		return nil, types.ErrSupplyInvariant.Wrap("burn amount exceeds recorded supply")
	}
	coins := sdk.NewCoins(msg.Amount)
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}
	afterSupply := m.bankKeeper.GetSupply(ctx, meta.Denom)
	if expected := beforeSupply.Sub(msg.Amount); !afterSupply.IsEqual(expected) {
		return nil, types.ErrSupplyInvariant.Wrapf("expected supply %s after burn, got %s", expected, afterSupply)
	}
	return &types.MsgBurnResponse{}, nil
}

func (m msgServer) ChangeAdmin(ctx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidDenom.Wrap("empty change admin request")
	}
	sender, err := parseAccAddress("sender", msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Denom)
	}
	if err := requireAdmin(meta, sender, "change admin"); err != nil {
		return nil, err
	}
	if msg.NewAdmin == "" {
		meta.Admin = ""
	} else {
		newAdmin, err := parseAccAddress("new_admin", msg.NewAdmin)
		if err != nil {
			return nil, err
		}
		meta.Admin = newAdmin.String()
	}
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
	return &types.MsgChangeAdminResponse{}, nil
}

func parseAccAddress(field, value string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(value)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid %s address: %v", field, err)
	}
	return addr, nil
}

func requireAdmin(meta types.DenomAuthorityMetadata, sender sdk.AccAddress, operation string) error {
	if meta.Admin == "" {
		return types.ErrUnauthorized.Wrap("denom admin is renounced")
	}
	if meta.Admin != sender.String() {
		return types.ErrUnauthorized.Wrapf("only denom admin can %s", operation)
	}
	return nil
}

func validateFactoryAmount(meta types.DenomAuthorityMetadata, amount sdk.Coin, operation string) error {
	if amount.Denom != meta.Denom {
		return types.ErrInvalidDenom.Wrapf("%s amount denom must match managed denom %s", operation, meta.Denom)
	}
	if !amount.IsValid() || !amount.IsPositive() {
		return types.ErrInvalidDenom.Wrapf("%s amount must be positive", operation)
	}
	return nil
}
