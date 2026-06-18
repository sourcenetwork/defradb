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

// Doc ID mapping keys bridge storage IDs, public DocIDs, and block CIDs.
//
// Public DocIDs are derived from the genesis composite CID, but the datastore
// needs a stable key before that CID exists. Document data is therefore written
// under a local short ID, and these systemstore keys record how that short ID
// maps to the public DocID once the genesis block has been materialized.
//
// Key shapes:
//   - /docid/{collectionShortID}/s/{shortDocID} -> public DocID
//   - /docid/n/p/{publicDocID} -> encoded {collectionShortID, shortDocID}
//   - /docid/b/{blockCID}/{collectionShortID}/{publicDocID} -> empty value
//
// The block-CID mapping is only for document-owned blocks: composite, field,
// delete, and encryption blocks. It lets CID-only paths such as P2P access
// checks and signature verification recover the public DocID. collectionShortID
// is included because identical block CIDs can appear in different collections.
//
// The path segments are intentionally short because these keys are persisted for
// every document and document block.
const (
	SHORT_ID_TO_DOC_ID  = "s"
	DOC_ID_TO_LOCAL_ID  = "p"
	NODE_DOC_ID_INDEX   = "n"
	BLOCK_CID_TO_DOC_ID = "b"
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

func stringSegment(value string) []byte {
	if value == "" {
		return nil
	}
	return []byte(value)
}

// ShortIDToDocIDKey maps a short doc ID to its public doc ID.
type ShortIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*ShortIDToDocIDKey)(nil)

func NewShortIDToDocIDKey(collectionShortID uint32, shortDocID uint32) ShortIDToDocIDKey {
	return ShortIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			collectionShortIDSegment(collectionShortID),
			[]byte(SHORT_ID_TO_DOC_ID),
			EncodeDocShortID(shortDocID),
		),
	}
}

// NodeDocIDToShortIDKey maps a public doc ID to this node's local collection/doc IDs.
type NodeDocIDToShortIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*NodeDocIDToShortIDKey)(nil)

func NewNodeDocIDToShortIDKey(docID string) NodeDocIDToShortIDKey {
	return NodeDocIDToShortIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			[]byte(NODE_DOC_ID_INDEX),
			[]byte(DOC_ID_TO_LOCAL_ID),
			stringSegment(docID),
		),
	}
}

// BlockCIDToDocIDKey records that a public DocID references a document-owned block CID.
type BlockCIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*BlockCIDToDocIDKey)(nil)

func NewBlockCIDToDocIDKey(collectionShortID uint32, blockCID string, publicDocID string) BlockCIDToDocIDKey {
	return BlockCIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			[]byte(BLOCK_CID_TO_DOC_ID),
			stringSegment(blockCID),
			collectionShortIDSegment(collectionShortID),
			stringSegment(publicDocID),
		),
	}
}
