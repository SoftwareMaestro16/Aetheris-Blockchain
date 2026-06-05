package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractNamespaceSetGetAndVersioning(t *testing.T) {
	store, err := NewStore(DefaultParams())
	require.NoError(t, err)
	key := Key{Namespace: ContractNamespace("alice"), Path: "counter"}
	require.NoError(t, store.Set(key, []byte{1}))
	entry, ok, err := store.Get(key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, uint64(1), entry.Version)
	require.Equal(t, []byte{1}, entry.Value)
	require.Equal(t, uint64(len(key.Namespace)+len(key.Path)+1), store.StateBytes())
	require.Equal(t, store.StateBytes()*DefaultStorageRentPerByte, store.StorageRent())
}

func TestBoundedIterationAndSnapshotImport(t *testing.T) {
	store, err := NewStore(DefaultParams())
	require.NoError(t, err)
	ns := ContractNamespace("alice")
	require.NoError(t, store.Set(Key{Namespace: ns, Path: "b"}, []byte{2}))
	require.NoError(t, store.Set(Key{Namespace: ns, Path: "a"}, []byte{1}))
	require.NoError(t, store.Set(Key{Namespace: ns, Path: "c"}, []byte{3}))

	entries, err := store.Iterate(ns, 2)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "a", entries[0].Key.Path)
	require.Equal(t, "b", entries[1].Key.Path)

	snapshot := store.Snapshot()
	imported, err := ImportSnapshot(DefaultParams(), snapshot)
	require.NoError(t, err)
	require.Equal(t, snapshot, imported.Snapshot())
}

func TestStorageRejectsOversizedAndMalformedState(t *testing.T) {
	params := DefaultParams()
	params.MaxStateBytes = 10
	store, err := NewStore(params)
	require.NoError(t, err)
	err = store.Set(Key{Namespace: "n", Path: "k"}, []byte(strings.Repeat("x", 20)))
	require.ErrorContains(t, err, "state size")

	err = store.Set(Key{Namespace: "", Path: "k"}, []byte{1})
	require.ErrorContains(t, err, "namespace")
	_, err = store.Iterate("n", 0)
	require.ErrorContains(t, err, "limit")
}

func TestSnapshotRootMismatchRejected(t *testing.T) {
	store, err := NewStore(DefaultParams())
	require.NoError(t, err)
	require.NoError(t, store.Set(Key{Namespace: "n", Path: "k"}, []byte{1}))
	snapshot := store.Snapshot()
	snapshot.StateRoot[0] ^= 0xff
	_, err = ImportSnapshot(DefaultParams(), snapshot)
	require.ErrorContains(t, err, "state root")
}
