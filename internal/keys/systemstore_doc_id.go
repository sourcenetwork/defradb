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

// Doc ID mapping keys bridge storage IDs and public DocIDs.
//
// Public DocIDs are derived from the genesis composite CID, but the datastore
// needs a stable key before that CID exists. Document data is therefore written
// under a local short ID, and these systemstore keys record how that short ID
// maps to the public DocID once the genesis block has been materialized.
//
// Collection-scoped mappings are used by normal document reads and writes.
// Node-scoped mappings support paths that only have a public DocID, such as P2P
// and block-signing lookups. Block indexes are stored in both directions:
// block CID -> DocID for verification, and DocID -> block CID for cleanup.
//
// The path segments are intentionally short because these keys are persisted for
// every document and document block.
const (
	SHORT_ID_TO_DOC_ID  = "s"
	DOC_ID_TO_SHORT_ID  = "p"
	NODE_DOC_ID_INDEX   = "n"
	BLOCK_CID_TO_DOC_ID = "b"
	DOC_ID_TO_BLOCK_CID = "pb"
)

type systemstoreDocIDKey struct {
	segments [][]byte
}

func newSystemstoreDocIDKey(segments ...[]byte) systemstoreDocIDKey {
	return systemstoreDocIDKey{segments: segments}
}

func (k systemstoreDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k systemstoreDocIDKey) Bytes() []byte {
	result := []byte(DOC_ID_INDEX)
	for _, segment := range k.segments {
		if len(segment) != 0 {
			result = append(result, '/')
			result = append(result, segment...)
		}
	}
	return result
}

func (k systemstoreDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func collectionShortIDSegment(collectionShortID uint32) []byte {
	if collectionShortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, uint64(collectionShortID))
}

func docIDSegment(docID string) []byte {
	if docID == "" {
		return nil
	}
	return []byte(docID)
}

// ShortIDToDocIDKey maps a short doc ID to its public doc ID.
type ShortIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*ShortIDToDocIDKey)(nil)

func NewShortIDToDocIDKey(collectionShortID uint32, shortDocID uint64) ShortIDToDocIDKey {
	return ShortIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			collectionShortIDSegment(collectionShortID),
			[]byte(SHORT_ID_TO_DOC_ID),
			EncodeDocShortID(shortDocID),
		),
	}
}

// DocIDToShortIDKey maps a public doc ID to its short doc ID.
type DocIDToShortIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*DocIDToShortIDKey)(nil)

func NewDocIDToShortIDKey(collectionShortID uint32, docID string) DocIDToShortIDKey {
	return DocIDToShortIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			collectionShortIDSegment(collectionShortID),
			[]byte(DOC_ID_TO_SHORT_ID),
			docIDSegment(docID),
		),
	}
}

// NodeDocIDToShortIDKey maps a public doc ID to this node's local short doc ID.
type NodeDocIDToShortIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*NodeDocIDToShortIDKey)(nil)

func NewNodeDocIDToShortIDKey(docID string) NodeDocIDToShortIDKey {
	return NodeDocIDToShortIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			[]byte(NODE_DOC_ID_INDEX),
			[]byte(DOC_ID_TO_SHORT_ID),
			docIDSegment(docID),
		),
	}
}

// NodeShortIDToDocIDKey maps this node's local short doc ID to its public doc ID.
type NodeShortIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*NodeShortIDToDocIDKey)(nil)

func NewNodeShortIDToDocIDKey(shortDocID uint64) NodeShortIDToDocIDKey {
	return NodeShortIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			[]byte(NODE_DOC_ID_INDEX),
			[]byte(SHORT_ID_TO_DOC_ID),
			EncodeDocShortID(shortDocID),
		),
	}
}

// BlockCIDToDocIDKey maps a document block CID to one public doc ID that links to it.
type BlockCIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*BlockCIDToDocIDKey)(nil)

func NewBlockCIDToDocIDKey(collectionShortID uint32, blockCID string, docID string) BlockCIDToDocIDKey {
	return BlockCIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			collectionShortIDSegment(collectionShortID),
			[]byte(BLOCK_CID_TO_DOC_ID),
			docIDSegment(blockCID),
			docIDSegment(docID),
		),
	}
}

// DocIDToBlockCIDKey maps a public doc ID to one of its document block CIDs.
type DocIDToBlockCIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*DocIDToBlockCIDKey)(nil)

func NewDocIDToBlockCIDKey(collectionShortID uint32, docID string, blockCID string) DocIDToBlockCIDKey {
	return DocIDToBlockCIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			collectionShortIDSegment(collectionShortID),
			[]byte(DOC_ID_TO_BLOCK_CID),
			docIDSegment(docID),
			docIDSegment(blockCID),
		),
	}
}
