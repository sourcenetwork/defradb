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
)

// Doc ID mapping keys bridge local storage references, DocIDs, and block CIDs.
//
// DocIDs are derived from the genesis composite CID, but the datastore
// needs a stable key before that CID exists. Document data is therefore written
// under a local short ID, and these systemstore keys record how that short ID
// maps to the DocID once the genesis block has been materialized.
//
// Key shapes:
//   - /d/s/{docShortID} -> DocID
//   - /d/p/{docID} -> encoded DocRef
//   - /d/r/{docShortID}/{docID} -> DocID
//   - /d/b/{blockCID} -> DocID
//
// The block-CID mapping is only for document-owned blocks: composite, field,
// delete, and encryption blocks. It lets CID-only paths such as P2P access
// checks and signature verification recover the DocID.
//
// The path segments are intentionally short because these keys are persisted for
// every document and document block.
const (
	SHORT_ID_TO_DOC_ID  = "s"
	DOC_ID_TO_DOC_REF   = "p"
	DOC_REF_TO_DOC_ID   = "r"
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

func stringSegment(value string) []byte {
	if value == "" {
		return nil
	}
	return []byte(value)
}

// ShortIDToDocIDKey maps a node-unique short doc ID to its DocID.
type ShortIDToDocIDKey struct {
	DocShortID uint64
}

var _ Key = (*ShortIDToDocIDKey)(nil)

func NewShortIDToDocIDKey(docShortID uint64) ShortIDToDocIDKey {
	return ShortIDToDocIDKey{
		DocShortID: docShortID,
	}
}

func (k ShortIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k ShortIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(SHORT_ID_TO_DOC_ID),
		EncodeDocShortID(k.DocShortID),
	)
}

func (k ShortIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// DocIDToDocRefKey maps a DocID to this node's local DocRef.
type DocIDToDocRefKey struct {
	DocID string
}

var _ Key = (*DocIDToDocRefKey)(nil)

func NewDocIDToDocRefKey(docID string) DocIDToDocRefKey {
	return DocIDToDocRefKey{
		DocID: docID,
	}
}

func (k DocIDToDocRefKey) ToString() string {
	return string(k.Bytes())
}

func (k DocIDToDocRefKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(DOC_ID_TO_DOC_REF),
		stringSegment(k.DocID),
	)
}

func (k DocIDToDocRefKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// DocRefToDocIDKey indexes all DocID aliases for a local doc short ID.
type DocRefToDocIDKey struct {
	DocShortID uint64
	DocID      string
}

var _ Key = (*DocRefToDocIDKey)(nil)

func NewDocRefToDocIDKey(docShortID uint64, docID string) DocRefToDocIDKey {
	return DocRefToDocIDKey{
		DocShortID: docShortID,
		DocID:      docID,
	}
}

func (k DocRefToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k DocRefToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(DOC_REF_TO_DOC_ID),
		EncodeDocShortID(k.DocShortID),
		stringSegment(k.DocID),
	)
}

func (k DocRefToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// BlockCIDToDocIDKey maps a document-owned block CID to the DocID that owns it.
type BlockCIDToDocIDKey struct {
	BlockCID string
}

var _ Key = (*BlockCIDToDocIDKey)(nil)

func NewBlockCIDToDocIDKey(blockCID string) BlockCIDToDocIDKey {
	return BlockCIDToDocIDKey{
		BlockCID: blockCID,
	}
}

func (k BlockCIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k BlockCIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(BLOCK_CID_TO_DOC_ID),
		stringSegment(k.BlockCID),
	)
}

func (k BlockCIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
