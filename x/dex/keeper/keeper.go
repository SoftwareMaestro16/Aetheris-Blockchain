package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sovereign-l1/l1/x/dex/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	bankKeeper   types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, bankKeeper types.BankKeeper) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, bankKeeper: bankKeeper}
}

func poolKey(id uint64) []byte {
	key := make([]byte, 1+8)
	key[0] = types.PoolPrefix[0]
	binary.BigEndian.PutUint64(key[1:], id)
	return key
}

func pairKey(denom0, denom1 string) []byte {
	key := make([]byte, 0, len(types.PairPrefix)+len(denom0)+1+len(denom1))
	key = append(key, types.PairPrefix...)
	key = append(key, []byte(denom0)...)
	key = append(key, 0)
	return append(key, []byte(denom1)...)
}

func parseInt(value string) sdkmath.Int {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.ZeroInt()
	}
	return out
}

func intString(value sdkmath.Int) string { return value.String() }

func parsePositiveInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok || !out.IsPositive() {
		return sdkmath.Int{}, types.ErrInvalidPool.Wrapf("%s must be a positive integer", field)
	}
	return out, nil
}

func parseNonNegativeInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok || out.IsNegative() {
		return sdkmath.Int{}, types.ErrInvalidPool.Wrapf("%s must be a non-negative integer", field)
	}
	return out, nil
}

func validatePoolState(pool types.Pool) (reserve0, reserve1, totalShares sdkmath.Int, err error) {
	if pool.Id == 0 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool id must be positive")
	}
	if err := sdk.ValidateDenom(pool.Denom0); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom0: %v", err)
	}
	if err := sdk.ValidateDenom(pool.Denom1); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom1: %v", err)
	}
	if pool.Denom0 >= pool.Denom1 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool denoms must be unique and canonical")
	}
	if pool.LpDenom != lpDenom(pool.Id) {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("invalid lp denom")
	}
	reserve0, err = parsePositiveInt("reserve0", pool.Reserve0)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	reserve1, err = parsePositiveInt("reserve1", pool.Reserve1)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	totalShares, err = parsePositiveInt("total_shares", pool.TotalShares)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	return reserve0, reserve1, totalShares, nil
}

func (k Keeper) GetNextPoolID(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.NextPoolIDKey)
	if err != nil || bz == nil {
		return types.DefaultNextPoolID, err
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) SetNextPoolID(ctx context.Context, id uint64) error {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return k.storeService.OpenKVStore(ctx).Set(types.NextPoolIDKey, bz)
}

func (k Keeper) SetPool(ctx context.Context, pool types.Pool) error {
	bz, err := k.cdc.Marshal(&pool)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(poolKey(pool.Id), bz)
}

func (k Keeper) GetPool(ctx context.Context, id uint64) (types.Pool, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(poolKey(id))
	if err != nil || bz == nil {
		return types.Pool{}, false, err
	}
	var pool types.Pool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		return types.Pool{}, false, err
	}
	return pool, true, nil
}

func (k Keeper) GetAllPools(ctx context.Context) ([]types.Pool, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.PoolPrefix, storetypes.PrefixEndBytes(types.PoolPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var pools []types.Pool
	for ; iter.Valid(); iter.Next() {
		var pool types.Pool
		if err := k.cdc.Unmarshal(iter.Value(), &pool); err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

func (k Keeper) GetPoolsPaginated(ctx context.Context, pageReq *querytypes.PageRequest) ([]types.Pool, *querytypes.PageResponse, error) {
	if pageReq == nil {
		pageReq = &querytypes.PageRequest{}
	}
	if len(pageReq.Key) > 0 && pageReq.Offset > 0 {
		return nil, nil, types.ErrInvalidPagination.Wrap("pagination key and offset cannot both be set")
	}
	if pageReq.CountTotal {
		return nil, nil, types.ErrInvalidPagination.Wrap("count_total is not supported for dex pools")
	}
	if pageReq.Reverse {
		return nil, nil, types.ErrInvalidPagination.Wrap("reverse pagination is not supported for dex pools")
	}
	if len(pageReq.Key) > 0 && !bytes.HasPrefix(pageReq.Key, types.PoolPrefix) {
		return nil, nil, types.ErrInvalidPagination.Wrap("pagination key must use dex pool prefix")
	}
	if pageReq.Offset > types.MaxPoolsQueryOffset {
		return nil, nil, types.ErrInvalidPagination.Wrapf("pagination offset cannot exceed %d", types.MaxPoolsQueryOffset)
	}

	limit := pageReq.Limit
	if limit == 0 {
		limit = types.DefaultPoolsQueryLimit
	}
	if limit > types.MaxPoolsQueryLimit {
		limit = types.MaxPoolsQueryLimit
	}

	start := types.PoolPrefix
	if len(pageReq.Key) > 0 {
		start = pageReq.Key
	}
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(start, storetypes.PrefixEndBytes(types.PoolPrefix))
	if err != nil {
		return nil, nil, err
	}
	defer iter.Close()

	pools := make([]types.Pool, 0, int(limit))
	var nextKey, lastKey []byte
	var skipped uint64

	for ; iter.Valid(); iter.Next() {
		key := append([]byte(nil), iter.Key()...)
		if len(pageReq.Key) > 0 && bytes.Compare(key, pageReq.Key) <= 0 {
			continue
		}
		if len(pageReq.Key) == 0 && skipped < pageReq.Offset {
			skipped++
			continue
		}
		if uint64(len(pools)) == limit {
			nextKey = lastKey
			break
		}

		var pool types.Pool
		if err := k.cdc.Unmarshal(iter.Value(), &pool); err != nil {
			return nil, nil, err
		}
		pools = append(pools, pool)
		lastKey = key
	}

	return pools, &querytypes.PageResponse{NextKey: nextKey}, nil
}

func (k Keeper) GetPoolIDByPair(ctx context.Context, denom0, denom1 string) (uint64, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(pairKey(denom0, denom1))
	if err != nil || bz == nil {
		return 0, false, err
	}
	if len(bz) != 8 {
		return 0, false, types.ErrInvalidPool.Wrap("invalid pair index value")
	}
	return binary.BigEndian.Uint64(bz), true, nil
}

func (k Keeper) SetPoolPairIndex(ctx context.Context, pool types.Pool) error {
	if _, _, _, err := validatePoolState(pool); err != nil {
		return err
	}
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, pool.Id)
	return k.storeService.OpenKVStore(ctx).Set(pairKey(pool.Denom0, pool.Denom1), bz)
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) {
	if gs.NextPoolId == 0 {
		gs.NextPoolId = types.DefaultNextPoolID
	}
	if err := gs.Validate(); err != nil {
		panic(err)
	}
	if err := k.SetNextPoolID(ctx, gs.NextPoolId); err != nil {
		panic(err)
	}
	for _, pool := range gs.Pools {
		if err := k.SetPool(ctx, pool); err != nil {
			panic(err)
		}
		if err := k.SetPoolPairIndex(ctx, pool); err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	next, err := k.GetNextPoolID(ctx)
	if err != nil {
		panic(err)
	}
	pools, err := k.GetAllPools(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{NextPoolId: next, Pools: pools}
}

func canonicalPair(a, b sdk.Coin) (sdk.Coin, sdk.Coin, error) {
	if !a.IsValid() || !b.IsValid() || !a.IsPositive() || !b.IsPositive() {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidLiquidity.Wrap("tokens must be positive")
	}
	if a.Denom == b.Denom {
		return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidPool.Wrap("pool requires two different denoms")
	}
	coins := []sdk.Coin{a, b}
	sort.Slice(coins, func(i, j int) bool { return coins[i].Denom < coins[j].Denom })
	return coins[0], coins[1], nil
}

func canonicalDenomPair(a, b string) (string, string, error) {
	if err := sdk.ValidateDenom(a); err != nil {
		return "", "", types.ErrInvalidPool.Wrapf("invalid denom_a: %v", err)
	}
	if err := sdk.ValidateDenom(b); err != nil {
		return "", "", types.ErrInvalidPool.Wrapf("invalid denom_b: %v", err)
	}
	if a == b {
		return "", "", types.ErrInvalidPool.Wrap("pool requires two different denoms")
	}
	if a > b {
		return b, a, nil
	}
	return a, b, nil
}

func coinsForPool(pool types.Pool, a, b sdk.Coin) (sdk.Coin, sdk.Coin, error) {
	if a.Denom == pool.Denom0 && b.Denom == pool.Denom1 {
		return a, b, nil
	}
	if a.Denom == pool.Denom1 && b.Denom == pool.Denom0 {
		return b, a, nil
	}
	return sdk.Coin{}, sdk.Coin{}, types.ErrInvalidLiquidity.Wrap("liquidity denoms do not match pool")
}

func lpDenom(poolID uint64) string {
	return fmt.Sprintf("%s/%d", types.LPDenomPrefix, poolID)
}

func minInt(a, b sdkmath.Int) sdkmath.Int {
	if a.LT(b) {
		return a
	}
	return b
}

func calcSwapOut(reserveIn, reserveOut, amountIn sdkmath.Int) sdkmath.Int {
	amountInAfterFee := amountIn.MulRaw(types.BpsDenominator - types.PoolFeeBps).QuoRaw(types.BpsDenominator)
	return reserveOut.Mul(amountInAfterFee).Quo(reserveIn.Add(amountInAfterFee))
}
