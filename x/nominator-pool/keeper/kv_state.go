package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

var (
	kvLayoutKey  = []byte("staking/meta/kv_layout")
	kvVersionKey = []byte("staking/meta/version")
	kvParamsKey  = []byte("staking/params")

	validatorPrefix                  = []byte("staking/validator/")
	validatorPerformanceScorePrefix  = []byte("staking/validator_score/")
	validatorCommissionPrefix        = []byte("staking/validator_commission/")
	validatorSlashingRiskPrefix      = []byte("staking/validator_slashing_risk/")
	validatorAllocationLimitPrefix   = []byte("staking/validator_allocation_limit/")
	poolPrefix                       = []byte("staking/pool/")
	liquidPoolPrefix                 = []byte("staking/liquid_pool/")
	poolSharePrefix                  = []byte("staking/pool_share/")
	poolValidatorAllocationPrefix    = []byte("staking/pool_allocation/")
	poolUnbondingPrefix              = []byte("staking/pool_unbonding/")
	poolRewardIndexPrefix            = []byte("staking/pool_reward_index/")
	rewardClaimPrefix                = []byte("staking/reward_claim/")
	stakeReputationAccumulatorPrefix = []byte("staking/reputation_accumulator/")
	identityReputationPrefix         = []byte("identity/reputation/")
	epochStakingSnapshotPrefix       = []byte("staking/snapshot/epoch/")
	validatorSetSnapshotPrefix       = []byte("staking/snapshot/validator_set/")
	validatorSlashEventPrefix        = []byte("staking/slash_event/")
)

func (k Keeper) loadKVGenesisState(ctx context.Context) (GenesisState, bool, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), false, nil
	}
	store := k.storeService.OpenKVStore(ctx)
	if legacy, err := store.Get(genesisKey); err != nil {
		return GenesisState{}, false, err
	} else if len(legacy) != 0 {
		var migrated GenesisState
		if err := json.Unmarshal(legacy, &migrated); err != nil {
			return GenesisState{}, false, err
		}
		migrated.State = migrated.State.Normalize(migrated.Params)
		if err := migrated.Validate(); err != nil {
			return GenesisState{}, false, err
		}
		if err := k.writeKVGenesisState(ctx, migrated); err != nil {
			return GenesisState{}, false, err
		}
		if err := store.Delete(genesisKey); err != nil {
			return GenesisState{}, false, err
		}
		return cloneGenesis(migrated), true, nil
	}
	marker, err := store.Get(kvLayoutKey)
	if err != nil || len(marker) == 0 {
		return DefaultGenesis(), false, err
	}
	versionBytes, err := store.Get(kvVersionKey)
	if err != nil {
		return GenesisState{}, false, err
	}
	paramsBytes, err := store.Get(kvParamsKey)
	if err != nil {
		return GenesisState{}, false, err
	}
	if len(versionBytes) == 0 || len(paramsBytes) == 0 {
		return GenesisState{}, false, fmt.Errorf("nominator pool KV genesis missing version or params")
	}
	var version uint64
	var params types.Params
	if err := json.Unmarshal(versionBytes, &version); err != nil {
		return GenesisState{}, false, err
	}
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return GenesisState{}, false, err
	}
	state, err := k.readKVState(ctx)
	if err != nil {
		return GenesisState{}, false, err
	}
	gs := GenesisState{Version: version, Params: params, State: state.Normalize(params)}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, false, err
	}
	return cloneGenesis(gs), true, nil
}

func (k Keeper) writeKVGenesisState(ctx context.Context, gs GenesisState) error {
	if k.storeService == nil {
		return nil
	}
	gs.State = gs.State.Normalize(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	for _, prefix := range nominatorPoolRecordPrefixes() {
		if err := deletePrefix(ctx, k, prefix); err != nil {
			return err
		}
	}
	if err := store.Set(kvLayoutKey, []byte("records-v1")); err != nil {
		return err
	}
	if err := setJSON(ctx, k, kvVersionKey, gs.Version); err != nil {
		return err
	}
	if err := setJSON(ctx, k, kvParamsKey, gs.Params); err != nil {
		return err
	}
	for key, value := range kvGenesisRecords(gs) {
		if err := store.Set([]byte(key), value); err != nil {
			return err
		}
	}
	return store.Delete(genesisKey)
}

func (k Keeper) writeKVGenesisDiff(ctx context.Context, before, after GenesisState) error {
	if k.storeService == nil {
		return nil
	}
	before.State = before.State.Normalize(before.Params)
	after.State = after.State.Normalize(after.Params)
	if err := after.Validate(); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(kvLayoutKey, []byte("records-v1")); err != nil {
		return err
	}
	if err := setJSON(ctx, k, kvVersionKey, after.Version); err != nil {
		return err
	}
	if err := setJSON(ctx, k, kvParamsKey, after.Params); err != nil {
		return err
	}
	oldRecords := kvGenesisRecords(before)
	newRecords := kvGenesisRecords(after)
	for key := range oldRecords {
		if _, found := newRecords[key]; !found {
			if err := store.Delete([]byte(key)); err != nil {
				return err
			}
		}
	}
	for key, value := range newRecords {
		if old, found := oldRecords[key]; found && bytes.Equal(old, value) {
			continue
		}
		if err := store.Set([]byte(key), value); err != nil {
			return err
		}
	}
	return store.Delete(genesisKey)
}

func (k Keeper) readKVState(ctx context.Context) (types.State, error) {
	var state types.State
	var err error
	if state.Pools, err = collectKVRecords[types.NominatorPool](ctx, k, poolPrefix); err != nil {
		return types.State{}, err
	}
	if state.Validators, err = collectKVRecords[types.Validator](ctx, k, validatorPrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorPerformanceScores, err = collectKVRecords[types.ValidatorPerformanceScore](ctx, k, validatorPerformanceScorePrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorCommissions, err = collectKVRecords[types.ValidatorCommission](ctx, k, validatorCommissionPrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorSlashingRisks, err = collectKVRecords[types.ValidatorSlashingRisk](ctx, k, validatorSlashingRiskPrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorAllocationLimits, err = collectKVRecords[types.ValidatorAllocationLimit](ctx, k, validatorAllocationLimitPrefix); err != nil {
		return types.State{}, err
	}
	if state.LiquidStakingPools, err = collectKVRecords[types.LiquidStakingPool](ctx, k, liquidPoolPrefix); err != nil {
		return types.State{}, err
	}
	if state.PoolShares, err = collectKVRecords[types.PoolShare](ctx, k, poolSharePrefix); err != nil {
		return types.State{}, err
	}
	if state.PoolValidatorAllocations, err = collectKVRecords[types.PoolValidatorAllocation](ctx, k, poolValidatorAllocationPrefix); err != nil {
		return types.State{}, err
	}
	if state.PoolUnbondingRequests, err = collectKVRecords[types.PoolUnbondingRequest](ctx, k, poolUnbondingPrefix); err != nil {
		return types.State{}, err
	}
	if state.PoolRewardIndexes, err = collectKVRecords[types.PoolRewardIndex](ctx, k, poolRewardIndexPrefix); err != nil {
		return types.State{}, err
	}
	if state.RewardClaims, err = collectKVRecords[types.RewardClaim](ctx, k, rewardClaimPrefix); err != nil {
		return types.State{}, err
	}
	if state.IdentityReputationRecords, err = collectKVRecords[types.IdentityReputationRecord](ctx, k, identityReputationPrefix); err != nil {
		return types.State{}, err
	}
	legacyReputation, err := collectKVRecords[types.StakeReputationAccumulator](ctx, k, stakeReputationAccumulatorPrefix)
	if err != nil {
		return types.State{}, err
	}
	state.IdentityReputationRecords = types.MergeIdentityReputationRecords(state.IdentityReputationRecords, legacyReputation)
	if state.EpochStakingSnapshots, err = collectKVRecords[types.EpochStakingSnapshot](ctx, k, epochStakingSnapshotPrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorSetSnapshots, err = collectKVRecords[types.ValidatorSetSnapshot](ctx, k, validatorSetSnapshotPrefix); err != nil {
		return types.State{}, err
	}
	if state.ValidatorSlashEvents, err = collectKVRecords[types.ValidatorSlashEvent](ctx, k, validatorSlashEventPrefix); err != nil {
		return types.State{}, err
	}
	return state, nil
}

func kvGenesisRecords(gs GenesisState) map[string][]byte {
	state := gs.State.Normalize(gs.Params)
	records := map[string][]byte{}
	put := func(key []byte, value any) {
		bz, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		records[string(key)] = bz
	}
	for _, value := range state.Pools {
		put(types.PoolKey(value.PoolID), value)
	}
	for _, value := range state.Validators {
		put(types.ValidatorKey(value.Address), value)
	}
	for _, value := range state.ValidatorPerformanceScores {
		put(validatorPerformanceScoreKey(value.Validator, value.Epoch), value)
	}
	for _, value := range state.ValidatorCommissions {
		put(validatorCommissionKey(value.Validator, value.Epoch), value)
	}
	for _, value := range state.ValidatorSlashingRisks {
		put(validatorSlashingRiskKey(value.Validator, value.Epoch), value)
	}
	for _, value := range state.ValidatorAllocationLimits {
		put(validatorAllocationLimitKey(value.Validator, value.Epoch), value)
	}
	for _, value := range state.LiquidStakingPools {
		put(liquidPoolKey(value.PoolID), value)
	}
	for _, value := range state.PoolShares {
		put(types.PoolShareKey(value.PoolID, value.Owner), value)
	}
	for _, value := range state.PoolValidatorAllocations {
		put(types.PoolAllocationKey(value.PoolID, value.Validator), value)
	}
	for _, value := range state.PoolUnbondingRequests {
		put(types.PoolUnbondingKey(value.PoolID, value.Owner, value.RequestID), value)
	}
	for _, value := range state.PoolRewardIndexes {
		put(types.PoolRewardIndexKey(value.PoolID), value)
	}
	for _, value := range state.RewardClaims {
		put(types.RewardClaimKey(value.PoolID, value.Owner, value.Epoch), value)
	}
	for _, value := range state.IdentityReputationRecords {
		put(types.IdentityReputationKey(value.Account), value)
	}
	for _, value := range state.EpochStakingSnapshots {
		put(types.EpochSnapshotKey(value.Epoch), value)
	}
	for _, value := range state.ValidatorSetSnapshots {
		put(types.ValidatorSetSnapshotKey(value.HeightOrEpoch), value)
	}
	for _, value := range state.ValidatorSlashEvents {
		put(validatorSlashEventKey(value), value)
	}
	return records
}

func collectKVRecords[T any](ctx context.Context, k Keeper, prefix []byte) ([]T, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	out := []T{}
	for ; iter.Valid(); iter.Next() {
		var value T
		if err := json.Unmarshal(iter.Value(), &value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

func setJSON(ctx context.Context, k Keeper, key []byte, value any) error {
	bz, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(key, bz)
}

func deletePrefix(ctx context.Context, k Keeper, prefix []byte) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}
	keys := [][]byte{}
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, append([]byte(nil), iter.Key()...))
	}
	if err := iter.Close(); err != nil {
		return err
	}
	for _, key := range keys {
		if err := store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

func nominatorPoolRecordPrefixes() [][]byte {
	return [][]byte{
		validatorPrefix,
		validatorPerformanceScorePrefix,
		validatorCommissionPrefix,
		validatorSlashingRiskPrefix,
		validatorAllocationLimitPrefix,
		poolPrefix,
		liquidPoolPrefix,
		poolSharePrefix,
		poolValidatorAllocationPrefix,
		poolUnbondingPrefix,
		poolRewardIndexPrefix,
		rewardClaimPrefix,
		stakeReputationAccumulatorPrefix,
		identityReputationPrefix,
		epochStakingSnapshotPrefix,
		validatorSetSnapshotPrefix,
		validatorSlashEventPrefix,
	}
}

func validatorPerformanceScoreKey(validator string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/validator_score/%s/%020d", validator, epoch))
}

func validatorCommissionKey(validator string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/validator_commission/%s/%020d", validator, epoch))
}

func validatorSlashingRiskKey(validator string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/validator_slashing_risk/%s/%020d", validator, epoch))
}

func validatorAllocationLimitKey(validator string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/validator_allocation_limit/%s/%020d", validator, epoch))
}

func liquidPoolKey(poolID string) []byte {
	return []byte("staking/liquid_pool/" + poolID)
}

func validatorSlashEventKey(event types.ValidatorSlashEvent) []byte {
	return []byte(fmt.Sprintf("staking/slash_event/%020d/%s/%s/%s/%020d", event.Height, event.Validator, event.PoolID, event.Fault, event.Epoch))
}
