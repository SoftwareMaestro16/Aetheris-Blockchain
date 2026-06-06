package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version uint64
	Params  types.Params
	State   types.ConfigState
}

type Keeper struct {
	genesis      GenesisState
	storeService corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version: prototype.CurrentGenesisVersion,
		Params:  types.DefaultParams(),
		State:   types.ConfigState{},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("config unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if k.storeService == nil {
		return errors.New("config persistent store is not configured")
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) saveGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if k.storeService == nil {
		return errors.New("config persistent store is not configured")
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	gs.State.Entries = types.SortedEntries(gs.State.Entries)
	return cloneGenesis(gs), nil
}

func (k *Keeper) UpdateParams(authority string, params types.Params) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.Params = params
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) UpdateParamsState(ctx context.Context, authority string, params types.Params) error {
	gs, err := k.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := gs.Params.Authorize(authority); err != nil {
		return err
	}
	gs.Params = params
	return k.saveGenesisState(ctx, gs)
}

func (k *Keeper) UpsertEntry(authority string, key string, value string, height int64) (types.ConfigEntry, error) {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return types.ConfigEntry{}, err
	}
	if height < 0 {
		return types.ConfigEntry{}, errors.New("config update height must be non-negative")
	}
	entries := types.SortedEntries(k.genesis.State.Entries)
	idx := sort.Search(len(entries), func(i int) bool {
		return entries[i].Key >= key
	})
	version := uint64(1)
	if idx < len(entries) && entries[idx].Key == key {
		version = entries[idx].Version + 1
	} else if uint32(len(entries)+1) > k.genesis.Params.MaxEntries {
		return types.ConfigEntry{}, errors.New("config entries limit reached")
	}
	entry := types.ConfigEntry{
		Key:           key,
		Value:         value,
		Owner:         authority,
		Version:       version,
		UpdatedHeight: height,
	}
	if err := entry.Validate(k.genesis.Params); err != nil {
		return types.ConfigEntry{}, err
	}
	if idx < len(entries) && entries[idx].Key == key {
		entries[idx] = entry
	} else {
		entries = append(entries, types.ConfigEntry{})
		copy(entries[idx+1:], entries[idx:])
		entries[idx] = entry
	}
	next := cloneGenesis(k.genesis)
	next.State.Entries = entries
	if err := next.Validate(); err != nil {
		return types.ConfigEntry{}, err
	}
	k.genesis = next
	return entry, nil
}

func (k Keeper) UpsertEntryState(ctx context.Context, authority string, key string, value string, height int64) (types.ConfigEntry, error) {
	gs, err := k.ExportGenesisState(ctx)
	if err != nil {
		return types.ConfigEntry{}, err
	}
	memory := NewKeeper()
	memory.genesis = gs
	entry, err := memory.UpsertEntry(authority, key, value, height)
	if err != nil {
		return types.ConfigEntry{}, err
	}
	return entry, k.saveGenesisState(ctx, memory.genesis)
}

func (k Keeper) Entry(key string) (types.ConfigEntry, bool, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return types.ConfigEntry{}, false, err
	}
	idx := sort.Search(len(state.Entries), func(i int) bool {
		return state.Entries[i].Key >= key
	})
	if idx >= len(state.Entries) || state.Entries[idx].Key != key {
		return types.ConfigEntry{}, false, nil
	}
	return state.Entries[idx], true, nil
}

func (k Keeper) Entries() ([]types.ConfigEntry, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return nil, err
	}
	return types.SortedEntries(state.Entries), nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	gs := m.keeper.ExportGenesis()
	return gs.Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = types.CloneState(gs.State)
	return gs
}
