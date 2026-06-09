package chunk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkMapBasic(t *testing.T) {
	m := NewMap(nil)

	key1 := []byte("alice_balance")
	val1, _ := NewBuilder().SetData([]byte{0x00, 0x64}, 16).Build() // 100

	key2 := []byte("bob_balance")
	val2, _ := NewBuilder().SetData([]byte{0x00, 0xC8}, 16).Build() // 200

	// 1. Put
	m1, err := m.Put(key1, val1)
	require.NoError(t, err)
	require.NotNil(t, m1.Root())

	m2, err := m1.Put(key2, val2)
	require.NoError(t, err)

	// 2. Get
	res1, err := m2.Get(key1)
	require.NoError(t, err)
	require.Equal(t, val1.Hash(), res1.Hash())

	res2, err := m2.Get(key2)
	require.NoError(t, err)
	require.Equal(t, val2.Hash(), res2.Hash())

	// 3. Update
	val1Updated, _ := NewBuilder().SetData([]byte{0x00, 0x96}, 16).Build() // 150
	m3, err := m2.Put(key1, val1Updated)
	require.NoError(t, err)

	res1Updated, _ := m3.Get(key1)
	require.Equal(t, val1Updated.Hash(), res1Updated.Hash())

	// 4. Delete
	m4, err := m3.Delete(key1)
	require.NoError(t, err)
	res1Deleted, _ := m4.Get(key1)
	require.Nil(t, res1Deleted)

	res2StillThere, _ := m4.Get(key2)
	require.NotNil(t, res2StillThere)
	require.Equal(t, val2.Hash(), res2StillThere.Hash())
}

func TestChunkMapPersistence(t *testing.T) {
	m := NewMap(nil)
	key := []byte("persistent")
	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()

	m1, _ := m.Put(key, val)
	root1 := m1.Root().Hash()

	// Update with same value should result in same root (because Chunk hashes are deterministic)
	m2, _ := m1.Put(key, val)
	root2 := m2.Root().Hash()
	require.Equal(t, root1, root2)

	// Different key should result in different root
	m3, _ := m1.Put([]byte("other"), val)
	root3 := m3.Root().Hash()
	require.NotEqual(t, root1, root3)
}

func TestChunkMapProof(t *testing.T) {
	m := NewMap(nil)
	key1 := []byte("key1")
	val1, _ := NewBuilder().SetData([]byte{1}, 8).Build()
	key2 := []byte("key2")
	val2, _ := NewBuilder().SetData([]byte{2}, 8).Build()

	m, _ = m.Put(key1, val1)
	m, _ = m.Put(key2, val2)

	proof, err := m.Prove(key1)
	require.NoError(t, err)
	require.Equal(t, m.Root().Hash(), proof.Hash(), "proof root hash must match actual root hash")

	// Verify proof: we should be able to Get(key1) from proof tree
	pm := NewMap(proof)
	res1, _ := pm.Get(key1)
	require.NotNil(t, res1)
	require.Equal(t, val1.Hash(), res1.Hash())

	// Get(key2) from proof tree should return a pruned chunk or nil if pruned deeply
	res2, _ := pm.Get(key2)
	if res2 != nil {
		require.Equal(t, TypePruned, res2.TypeTag(), "key2 should be pruned in key1's proof")
	}
}
