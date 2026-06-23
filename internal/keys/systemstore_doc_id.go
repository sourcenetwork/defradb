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

// Doc ID mapping keys bridge local storage references, DocIDs, and block CIDs.
//
// DocIDs are derived from the genesis composite CID, but the datastore
// needs a stable key before that CID exists. Document data is therefore written
// under a local short ID, and these systemstore keys record how that short ID
// maps to the DocID once the genesis block has been materialized.
//
// Key shapes:
//   - /d/s/{collectionShortID}/{docShortID} -> DocID
//   - /d/p/{docID} -> encoded DocRef
//   - /d/b/{blockCID}/{collectionShortID} -> DocID
//
// The block-CID mapping is only for document-owned blocks: composite, field,
// delete, and encryption blocks. It lets CID-only paths such as P2P access
// checks and signature verification recover the DocID. collectionShortID is
// included because identical block CIDs can appear in different collections.
//
// collectionShortID is also part of the short-ID mapping because docShortID
// values are collection-scoped, not node-scoped. For the same reason, it is
// stored in DocRef values instead of being normalized behind a separate
// docShortID -> collectionShortID lookup.
//
// The path segments are intentionally short because these keys are persisted for
// every document and document block.
const (
	SHORT_ID_TO_DOC_ID  = "s"
	DOC_ID_TO_LOCAL_ID  = "p"
	BLOCK_CID_TO_DOC_ID = "b"
)

func newDocIDSystemstoreKey(segments ...[]byte) []byte {
	result := []byte(DOC_ID_INDEX)
	for _, segment := range segments {
		if len(segment) != 0 {
			result = append(result, '/')
			result = append(result, segment...)
		}
	}
	return result
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

// ShortIDToDocIDKey maps a collection-scoped short doc ID to its DocID.
type ShortIDToDocIDKey struct {
	CollectionShortID uint32
	DocShortID        uint32
}

var _ Key = (*ShortIDToDocIDKey)(nil)

func NewShortIDToDocIDKey(collectionShortID uint32, docShortID uint32) ShortIDToDocIDKey {
	return ShortIDToDocIDKey{
		CollectionShortID: collectionShortID,
		DocShortID:        docShortID,
	}
}

func (k ShortIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k ShortIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(SHORT_ID_TO_DOC_ID),
		collectionShortIDSegment(k.CollectionShortID),
		EncodeDocShortID(k.DocShortID),
	)
}

func (k ShortIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// NodeDocIDToShortIDKey maps a DocID to this node's local DocRef.
type NodeDocIDToShortIDKey struct {
	DocID string
}

var _ Key = (*NodeDocIDToShortIDKey)(nil)

func NewNodeDocIDToShortIDKey(docID string) NodeDocIDToShortIDKey {
	return NodeDocIDToShortIDKey{
		DocID: docID,
	}
}

func (k NodeDocIDToShortIDKey) ToString() string {
	return string(k.Bytes())
}

func (k NodeDocIDToShortIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(DOC_ID_TO_LOCAL_ID),
		stringSegment(k.DocID),
	)
}

func (k NodeDocIDToShortIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// BlockCIDToDocIDKey maps a document-owned block CID to the DocID that owns it.
type BlockCIDToDocIDKey struct {
	BlockCID          string
	CollectionShortID uint32
}

var _ Key = (*BlockCIDToDocIDKey)(nil)

func NewBlockCIDToDocIDKey(collectionShortID uint32, blockCID string) BlockCIDToDocIDKey {
	return BlockCIDToDocIDKey{
		BlockCID:          blockCID,
		CollectionShortID: collectionShortID,
	}
}

func (k BlockCIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k BlockCIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(BLOCK_CID_TO_DOC_ID),
		stringSegment(k.BlockCID),
		collectionShortIDSegment(k.CollectionShortID),
	)
}

func (k BlockCIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
