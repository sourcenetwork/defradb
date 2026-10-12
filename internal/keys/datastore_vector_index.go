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

// A vector index keeps three kinds of entry under its (collectionShortID, indexID, epoch) prefix: one
// meta entry, many node entries, and (for a deferred-maintenance index) many pending entries. These
// bytes mark which is which. A scalar index reuses the same
// prefix but stores field values there, which are never encoded as one of these marker bytes, so
// the two kinds of key never clash.
const (
	vectorIndexMetaDiscriminator byte = 'm'
	vectorIndexNodeDiscriminator byte = 'n'
	// vectorIndexPendingDiscriminator marks a pending entry: a write not yet applied to the graph.
	vectorIndexPendingDiscriminator byte = 'p'
)

// VectorIndexKeyKind says which kind of entry a VectorIndexKey addresses.
type VectorIndexKeyKind uint8

const (
	// VectorIndexKeyNode addresses a graph node entry. It is the zero value.
	VectorIndexKeyNode VectorIndexKeyKind = iota
	// VectorIndexKeyMeta addresses the graph's meta singleton.
	VectorIndexKeyMeta
	// VectorIndexKeyPending addresses a pending (not yet applied) write of one node.
	VectorIndexKeyPending
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
	// Kind is which kind of entry this key addresses.
	Kind VectorIndexKeyKind
	// NodeID is the node this key addresses. Only meaningful for node and pending keys. Node ids come from
	// a sequence that starts at 1, so a zero NodeID never addresses a node: it marks a node-prefix key
	// (`.../n`, or `.../p` for pending) that scans every entry of that kind, as opposed to a full key (`.../n/<id>`). This
	// matches how the other keys use a zero component as their prefix.
	NodeID uint64
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
		Kind:              VectorIndexKeyMeta,
	}
}

// NewVectorNodeKey returns the key of a single node entry in the vector index's graph.
func NewVectorNodeKey(collShortID, indexID, epoch uint32, nodeID uint64) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Kind:              VectorIndexKeyNode,
		NodeID:            nodeID,
	}
}

// NewVectorNodePrefix returns a key with no NodeID component (a zero NodeID), whose bytes form the
// common prefix of every node entry of the given (collection, index, epoch). It is intended for
// prefix iteration, not for a Get/Set/Delete of a single entry.
func NewVectorNodePrefix(collShortID, indexID, epoch uint32) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Kind:              VectorIndexKeyNode,
	}
}

// NewVectorPendingKey returns the key of the pending entry for one node of a deferred index.
func NewVectorPendingKey(collShortID, indexID, epoch uint32, nodeID uint64) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Kind:              VectorIndexKeyPending,
		NodeID:            nodeID,
	}
}

// NewVectorPendingPrefix returns a key whose bytes form the common prefix of every pending entry of
// the given (collection, index, epoch). It is intended for prefix iteration only.
func NewVectorPendingPrefix(collShortID, indexID, epoch uint32) VectorIndexKey {
	return VectorIndexKey{
		CollectionShortID: collShortID,
		IndexID:           indexID,
		Epoch:             epoch,
		Kind:              VectorIndexKeyPending,
	}
}

// DecodeVectorPendingNodeID returns the node id of a full pending key's bytes, as produced by
// Bytes on a NewVectorPendingKey.
func DecodeVectorPendingNodeID(b []byte) (uint64, error) {
	rest := b
	// Skip the three "/<uvarint>" components: collection, index and epoch.
	for range 3 {
		if len(rest) == 0 || rest[0] != '/' {
			return 0, ErrInvalidVectorPendingKey
		}
		var err error
		rest, _, err = encoding.DecodeUvarintAscending(rest[1:])
		if err != nil {
			return 0, err
		}
	}
	if len(rest) < 3 || rest[0] != '/' || rest[1] != vectorIndexPendingDiscriminator || rest[2] != '/' {
		return 0, ErrInvalidVectorPendingKey
	}
	_, nodeID, err := encoding.DecodeUvarintAscending(rest[3:])
	return nodeID, err
}

// Bytes returns the byte representation of the key.
//
// Layout:
//
//	/<collShortID>/<indexID>/<epoch>/m               (meta key)
//	/<collShortID>/<indexID>/<epoch>/n/<nodeID>      (node key)
//	/<collShortID>/<indexID>/<epoch>/n               (node prefix; no trailing nodeID)
//	/<collShortID>/<indexID>/<epoch>/p/<nodeID>      (pending key)
//	/<collShortID>/<indexID>/<epoch>/p               (pending prefix; no trailing nodeID)
func (k *VectorIndexKey) Bytes() []byte {
	b := encoding.EncodeUvarintAscending([]byte{'/'}, uint64(k.CollectionShortID))

	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.IndexID))

	b = append(b, '/')
	b = encoding.EncodeUvarintAscending(b, uint64(k.Epoch))

	b = append(b, '/')
	switch k.Kind {
	case VectorIndexKeyMeta:
		b = append(b, vectorIndexMetaDiscriminator)
	default:
		if k.Kind == VectorIndexKeyPending {
			b = append(b, vectorIndexPendingDiscriminator)
		} else {
			b = append(b, vectorIndexNodeDiscriminator)
		}
		if k.NodeID != 0 {
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
		Kind:              k.Kind,
		NodeID:            k.NodeID,
		offset:            k.offset + 1,
	}
}
