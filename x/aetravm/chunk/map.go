package chunk

import (
	"lukechampine.com/blake3"
)

// Map represents a persistent, immutable Merkle Trie with 8-fanout.
type Map struct {
	root *Chunk
}

func NewMap(root *Chunk) *Map {
	return &Map{root: root}
}

func (m *Map) Root() *Chunk {
	return m.root
}

// Get retrieves the value Chunk for the given key.
func (m *Map) Get(key []byte) (*Chunk, error) {
	if m.root == nil {
		return nil, nil
	}
	hash := HashKey(key)
	return m.getRecursive(m.root, hash[:], 0)
}

func (m *Map) getRecursive(node *Chunk, hash []byte, bitOffset int) (*Chunk, error) {
	if bitOffset >= 256 {
		// Terminal node
		return node, nil
	}

	index := getIndex(hash, bitOffset)
	child := node.RefAt(index)
	if child == nil {
		return nil, nil
	}

	return m.getRecursive(child, hash, bitOffset+3)
}

// Put inserts or updates a key-value pair, returning a new Map with the updated root.
func (m *Map) Put(key []byte, value *Chunk) (*Map, error) {
	hash := HashKey(key)
	newRoot, err := m.putRecursive(m.root, hash[:], 0, value)
	if err != nil {
		return nil, err
	}
	return &Map{root: newRoot}, nil
}

func (m *Map) putRecursive(node *Chunk, hash []byte, bitOffset int, value *Chunk) (*Chunk, error) {
	if bitOffset >= 256 {
		// Replace the terminal node
		return value, nil
	}

	index := getIndex(hash, bitOffset)

	builder := NewBuilder().SetTypeTag(TypeNormal)
	var childToUpdate *Chunk
	if node != nil {
		builder.SetData(node.Data(), node.BitCount())
		for i := 0; i < MaxRefs; i++ {
			builder.SetRef(i, node.RefAt(i))
		}
		childToUpdate = node.RefAt(index)
	}

	newChild, err := m.putRecursive(childToUpdate, hash, bitOffset+3, value)
	if err != nil {
		return nil, err
	}
	builder.SetRef(index, newChild)

	return builder.Build()
}

// Delete removes a key, returning a new Map.
func (m *Map) Delete(key []byte) (*Map, error) {
	if m.root == nil {
		return m, nil
	}
	hash := HashKey(key)
	newRoot, err := m.deleteRecursive(m.root, hash[:], 0)
	if err != nil {
		return nil, err
	}
	return &Map{root: newRoot}, nil
}

func (m *Map) deleteRecursive(node *Chunk, hash []byte, bitOffset int) (*Chunk, error) {
	if node == nil {
		return nil, nil
	}
	if bitOffset >= 256 {
		return nil, nil
	}

	index := getIndex(hash, bitOffset)
	child := node.RefAt(index)
	if child == nil {
		return node, nil
	}

	newChild, err := m.deleteRecursive(child, hash, bitOffset+3)
	if err != nil {
		return nil, err
	}

	// Check if the node becomes empty
	hasOtherChildren := false
	for i := 0; i < MaxRefs; i++ {
		if i == index {
			if newChild != nil {
				hasOtherChildren = true
			}
		} else {
			if node.RefAt(i) != nil {
				hasOtherChildren = true
			}
		}
	}

	if !hasOtherChildren && node.BitCount() == 0 {
		return nil, nil
	}

	builder := NewBuilder().SetTypeTag(TypeNormal).SetData(node.Data(), node.BitCount())
	for i := 0; i < MaxRefs; i++ {
		if i == index {
			builder.SetRef(i, newChild)
		} else {
			builder.SetRef(i, node.RefAt(i))
		}
	}
	return builder.Build()
}

// Prove returns a pruned Chunk tree that serves as a Merkle proof for the key.
func (m *Map) Prove(key []byte) (*Chunk, error) {
	if m.root == nil {
		return nil, nil
	}
	hash := HashKey(key)
	return m.proveRecursive(m.root, hash[:], 0)
}

func (m *Map) proveRecursive(node *Chunk, hash []byte, bitOffset int) (*Chunk, error) {
	if node == nil {
		return nil, nil
	}
	if node.TypeTag() == TypePruned {
		return node, nil
	}
	if bitOffset >= 256 {
		return node, nil
	}

	index := getIndex(hash, bitOffset)
	builder := NewBuilder().
		SetTypeTag(node.TypeTag()).
		SetData(node.Data(), node.BitCount())

	for i := 0; i < MaxRefs; i++ {
		child := node.RefAt(i)
		if child == nil {
			continue
		}
		if i == index {
			// Stay unpruned for the target path
			provedChild, err := m.proveRecursive(child, hash, bitOffset+3)
			if err != nil {
				return nil, err
			}
			builder.SetRef(i, provedChild)
		} else {
			// Prune other branches
			pruned, _ := NewPrunedChunk(child.Level(), [][]byte{child.HashLayer(0), child.HashLayer(1)})
			builder.SetRef(i, pruned)
		}
	}

	return builder.Build()
}

// HashKey computes the BLAKE3 hash of the key.
func HashKey(key []byte) [32]byte {
	return blake3.Sum256(key)
}

// getIndex extracts 3 bits from the hash at the given bit offset.
func getIndex(hash []byte, bitOffset int) int {
	// We treat the 256-bit hash as a bit stream.
	// We can safely read up to 258 bits by assuming trailing zeros.

	byteIdx := bitOffset / 8
	bitIdx := bitOffset % 8

	var val uint32
	val = uint32(hash[byteIdx]) << 16
	if byteIdx+1 < len(hash) {
		val |= uint32(hash[byteIdx+1]) << 8
	}
	if byteIdx+2 < len(hash) {
		val |= uint32(hash[byteIdx+2])
	}

	// We want 3 bits starting at bitIdx in the first byte (which is at bits 23..16 of val)
	// So bitIdx=0 means bits 23, 22, 21.
	// Shift is 24 - bitIdx - 3
	shift := 24 - bitIdx - 3
	return int((val >> shift) & 0x07)
}
