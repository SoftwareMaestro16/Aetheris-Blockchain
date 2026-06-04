package keeper

import (
	"bytes"
	"context"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	bankKeeper   types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper}
}

func denomKey(denom string) []byte {
	key := make([]byte, 0, len(types.DenomPrefix)+len(denom))
	key = append(key, types.DenomPrefix...)
	return append(key, []byte(denom)...)
}

func (k Keeper) SetDenom(ctx context.Context, meta types.DenomAuthorityMetadata) error {
	normalized, err := types.NormalizeDenomAuthorityMetadata(meta)
	if err != nil {
		return err
	}
	bz, err := k.cdc.Marshal(&normalized)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(denomKey(normalized.Denom), bz)
}

func (k Keeper) GetDenom(ctx context.Context, denom string) (types.DenomAuthorityMetadata, bool, error) {
	if err := types.ValidateFactoryDenom(denom); err != nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(denomKey(denom))
	if err != nil || bz == nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	var meta types.DenomAuthorityMetadata
	if err := k.cdc.Unmarshal(bz, &meta); err != nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	meta, err = types.NormalizeDenomAuthorityMetadata(meta)
	if err != nil {
		return types.DenomAuthorityMetadata{}, false, err
	}
	return meta, true, nil
}

func (k Keeper) GetAllDenoms(ctx context.Context) ([]types.DenomAuthorityMetadata, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.DenomPrefix, storetypes.PrefixEndBytes(types.DenomPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var out []types.DenomAuthorityMetadata
	for ; iter.Valid(); iter.Next() {
		var meta types.DenomAuthorityMetadata
		if err := k.cdc.Unmarshal(iter.Value(), &meta); err != nil {
			return nil, err
		}
		var err error
		meta, err = types.NormalizeDenomAuthorityMetadata(meta)
		if err != nil {
			return nil, err
		}
		out = append(out, meta)
	}
	return out, nil
}

func (k Keeper) FullDenom(creator, subdenom string) (string, error) {
	return types.BuildFactoryDenom(creator, subdenom)
}

func (k Keeper) GetDenomsPaginated(ctx context.Context, pageReq *querytypes.PageRequest) ([]types.DenomAuthorityMetadata, *querytypes.PageResponse, error) {
	if pageReq == nil {
		pageReq = &querytypes.PageRequest{}
	}
	if len(pageReq.Key) > 0 && pageReq.Offset > 0 {
		return nil, nil, types.ErrInvalidPagination.Wrap("pagination key and offset cannot both be set")
	}
	if pageReq.Reverse {
		return nil, nil, types.ErrInvalidPagination.Wrap("reverse pagination is not supported for tokenfactory denoms")
	}
	if len(pageReq.Key) > 0 && !bytes.HasPrefix(pageReq.Key, types.DenomPrefix) {
		return nil, nil, types.ErrInvalidPagination.Wrap("pagination key must use tokenfactory denom prefix")
	}

	limit := pageReq.Limit
	if limit == 0 {
		limit = types.DefaultDenomQueryLimit
	}
	if limit > types.MaxDenomQueryLimit {
		limit = types.MaxDenomQueryLimit
	}

	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.DenomPrefix, storetypes.PrefixEndBytes(types.DenomPrefix))
	if err != nil {
		return nil, nil, err
	}
	defer iter.Close()

	var (
		out       []types.DenomAuthorityMetadata
		nextKey   []byte
		lastKey   []byte
		total     uint64
		skipped   uint64
		pageFull  bool
		needTotal = pageReq.CountTotal && len(pageReq.Key) == 0
	)

	for ; iter.Valid(); iter.Next() {
		key := append([]byte(nil), iter.Key()...)
		if len(pageReq.Key) > 0 && bytes.Compare(key, pageReq.Key) <= 0 {
			continue
		}
		if needTotal {
			total++
		}
		if len(pageReq.Key) == 0 && skipped < pageReq.Offset {
			skipped++
			continue
		}
		if pageFull {
			continue
		}
		if uint64(len(out)) == limit {
			nextKey = lastKey
			pageFull = true
			if !needTotal {
				break
			}
			continue
		}

		var meta types.DenomAuthorityMetadata
		if err := k.cdc.Unmarshal(iter.Value(), &meta); err != nil {
			return nil, nil, err
		}
		meta, err = types.NormalizeDenomAuthorityMetadata(meta)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, meta)
		lastKey = key
	}

	return out, &querytypes.PageResponse{NextKey: nextKey, Total: total}, nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(err)
	}
	for _, meta := range gs.Denoms {
		if err := k.SetDenom(ctx, meta); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	denoms, err := k.GetAllDenoms(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{Denoms: denoms}
}

func BankMetadata(denom string) banktypes.Metadata {
	return banktypes.Metadata{
		Base:        denom,
		Display:     denom,
		Name:        denom,
		Symbol:      denom,
		Description: "factory token " + denom,
		DenomUnits:  []*banktypes.DenomUnit{{Denom: denom, Exponent: 0}},
	}
}
