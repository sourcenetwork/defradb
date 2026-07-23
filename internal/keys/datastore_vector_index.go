// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

// A vector index keeps two kinds of entry under its (collectionShortID, indexID, epoch) prefix: one
// meta entry and many node entries. These bytes mark which is which. A scalar index reuses the same
// prefix but stores field values there, which are never encoded as one of these marker bytes, so
// the two kinds of key never clash.
const (
	vectorIndexMetaDiscriminator byte = 'm'
	vectorIndexNodeDiscriminator byte = 'n'
)

// VectorIndexKey points to one entry in a vector index's graph: either the meta entry or a single
// node. Entries are grouped by (CollectionShortID, IndexID, Epoch), the same way scalar index keys
// are, so a rebuild's new epoch never overlaps the live one.
type VectorIndexKey struct {
	// CollectionShortID is the id of the collection.
	CollectionShortID uint32
	// IndexID is the id of the vector index.
	IndexID uint32
	// Epoch namespaces the entries of a single index build; a rebuild fills a fresh epoch
	// disjoint from the current one.
	Epoch uint32
	// IsMeta is true for the graph's meta singleton key, false for a node entry key.
	IsMeta bool
	// NodeID is the node this key addresses. Only meaningful when IsMeta is false.
	NodeID uint64
	// hasNodeID distinguishes a node key (NodeID set) from a node-prefix key (no NodeID; used to
	// iterate over every node entry of the index).
	hasNodeID bool
	// offset controls how many times bytesPrefixEnd is applied during encoding; see
	// IndexDataStoreKey.Offset for the same mechanism.
	offset uint64
}

var _ Walkable = (*VectorIndexKey)(nil)
var _ CollectionedKey = (*VectorIndexKey)(nil)

// NewVectorMetaKey returns the key of the vector index's meta singleton.
func NewVectorMetaKey(collShortID, indexID, epoch uint32) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		IsMeta:            true,
	}
}

// NewVectorNodeKey returns the key of a single node entry in the vector index's graph.
func NewVectorNodeKey(collShortID, indexID, epoch uint32, nodeID uint64) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		IsMeta:            false,
		NodeID:            nodeID,
		hasNodeID:         true,
	}
}

// NewVectorNodePrefix returns a key with no NodeID component, whose bytes form the common prefix
// of every node entry of the given (collection, index, epoch). It is intended for prefix
// iteration, not for a Get/Set/Delete of a single entry.
func NewVectorNodePrefix(collShortID, indexID, epoch uint32) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		IsMeta:            false,
		hasNodeID:         false,
	}
}

// Bytes returns the byte representation of the key.
//
// Layout:
//
//	/<collShortID>/<indexID>/<epoch>/m               (meta key)
//	/<collShortID>/<indexID>/<epoch>/n/<nodeID>      (node key)
//	/<collShortID>/<indexID>/<epoch>/n               (node prefix; no trailing nodeID)
func (k *VectorIndexKey) Bytes() []byte {
	b := encoding.EncodeUvarintAscending([]byte{'/'}, uint64(k.CollectionShortID))

	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.IndexID))

	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.Epoch))

	b = append(b, '/')
	if k.IsMeta {
		b = append(b, vectorIndexMetaDiscriminator)
	} else {
		b = append(b, vectorIndexNodeDiscriminator)
		if k.hasNodeID {
			b = append(b, '/')
			b = encoding.EncodeUvarintAscending(b, k.NodeID)
		}
	}

	for i := uint64(0); i < k.offset; i++ {
		b = bytesPrefixEnd(b)
	}

	return b
}

// ToString returns the string representation of the key.
func (k *VectorIndexKey) ToString() string {
	return string(k.Bytes())
}

// ToDS returns the datastore key.
func (k *VectorIndexKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// GetCollectionShortID returns the id of the collection this key belongs to.
func (k *VectorIndexKey) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}

// PrefixEnd returns a key that would sort immediately after all keys with this prefix.
// It returns a key such that all keys with the prefix are >= k and < k.PrefixEnd().
func (k *VectorIndexKey) PrefixEnd() Walkable {
	return &VectorIndexKey{
		CollectionShortID: k.CollectionShortID,
		IndexID:           k.IndexID,
		Epoch:             k.Epoch,
		IsMeta:            k.IsMeta,
		NodeID:            k.NodeID,
		hasNodeID:         k.hasNodeID,
		offset:            k.offset + 1,
	}
}
