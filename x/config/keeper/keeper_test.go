package keeper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestGenesisRejectsMalformedAndNondeterministicState(t *testing.T) {
	keeper := NewKeeper()

	bad := DefaultGenesis()
	bad.Version = 99
	require.ErrorContains(t, keeper.InitGenesis(bad), "unsupported")

	bad = DefaultGenesis()
	bad.Params.Authority = "4:0000000000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, keeper.InitGenesis(bad), "zero address")

	bad = DefaultGenesis()
	bad.State.Entries = []types.ConfigEntry{
		entry("runtime/z", "one", 1),
		entry("runtime/a", "two", 1),
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "sorted")

	bad.State.Entries = []types.ConfigEntry{
		entry("runtime/a", "one", 1),
		entry("runtime/a", "two", 1),
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "duplicated")
}

func TestUpsertRequiresAuthorityAndRejectsUnsafeFields(t *testing.T) {
	keeper := NewKeeper()

	_, err := keeper.UpsertEntry("4:0000000000000000000000000000000000000000000000000000000000000002", "runtime/max_validators", "100", 1)
	require.ErrorContains(t, err, "governance authority")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, " runtime/max_validators", "100", 1)
	require.ErrorContains(t, err, "canonical")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/max_validators", strings.Repeat("x", int(types.MaxConfigValueBytesV1)+1), 1)
	require.ErrorContains(t, err, "value exceeds")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/max_validators", "100", -1)
	require.ErrorContains(t, err, "height")
}

func TestUpsertMaintainsSortedEntriesAndVersions(t *testing.T) {
	keeper := NewKeeper()

	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/z", "z", 1)
	require.NoError(t, err)
	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 2)
	require.NoError(t, err)
	updated, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a2", 3)
	require.NoError(t, err)
	require.Equal(t, uint64(2), updated.Version)

	entries, err := keeper.Entries()
	require.NoError(t, err)
	require.Equal(t, []string{"runtime/a", "runtime/z"}, []string{entries[0].Key, entries[1].Key})

	found, ok, err := keeper.Entry("runtime/a")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "a2", found.Value)
	require.Equal(t, int64(3), found.UpdatedHeight)
}

func TestUpdateParamsAndEntryLimit(t *testing.T) {
	keeper := NewKeeper()
	params := types.DefaultParams()
	params.MaxEntries = 1
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, params))

	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 1)
	require.NoError(t, err)
	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/b", "b", 2)
	require.ErrorContains(t, err, "limit")
}

func TestExportImportDeterministicAndMigration(t *testing.T) {
	source := NewKeeper()
	_, err := source.UpsertEntry(prototype.DefaultAuthority, "runtime/b", "b", 2)
	require.NoError(t, err)
	_, err = source.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 1)
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Equal(t, []string{"runtime/a", "runtime/b"}, []string{exported.State.Entries[0].Key, exported.State.Entries[1].Key})

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func entry(key string, value string, version uint64) types.ConfigEntry {
	return types.ConfigEntry{
		Key:           key,
		Value:         value,
		Owner:         prototype.DefaultAuthority,
		Version:       version,
		UpdatedHeight: 1,
	}
}
