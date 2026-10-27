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

// Doc ID mapping keys bridge storage references, DocIDs, and block CIDs.
//
// DocIDs are derived from the genesis composite CID, but the datastore
// needs a stable key before that CID exists. Document data is therefore written
// under a document short ID, and these systemstore keys record how that short ID
// maps to the DocID once the genesis block has been materialized.
//
// Key shapes:
//   - /d/s/{docShortID} -> DocID
//   - /d/p/{docID} -> encoded DocRef
//   - /d/r/{docShortID}/{docID} -> DocID
//   - /d/b/{blockCID}/{docID} -> {}
//
// The block-CID mapping is only for document-owned blocks: composite, field,
// delete, and encryption blocks. It lets CID-only paths such as P2P access
// checks and signature verification recover the DocID.
//
// The path segments are intentionally short because these keys are persisted for
// every document and document block.
const (
	DOC_SHORT_ID_TO_DOC_ID       = "s"
	DOC_ID_TO_DOC_REF            = "p"
	DOC_SHORT_ID_TO_DOC_ID_ALIAS = "r"
	BLOCK_CID_TO_DOC_ID          = "b"
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

// DocShortIDToDocIDKey maps a node-unique document short ID to its DocID.
type DocShortIDToDocIDKey struct {
	DocShortID uint64
}

var _ Key = (*DocShortIDToDocIDKey)(nil)

func NewDocShortIDToDocIDKey(docShortID uint64) DocShortIDToDocIDKey {
	return DocShortIDToDocIDKey{
		DocShortID: docShortID,
	}
}

func (k DocShortIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k DocShortIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(DOC_SHORT_ID_TO_DOC_ID),
		EncodeDocShortID(k.DocShortID),
	)
}

func (k DocShortIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// DocIDToDocRefKey maps a DocID to this node's DocRef.
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

// DocShortIDToDocIDAliasKey indexes all DocID aliases for a document short ID.
type DocShortIDToDocIDAliasKey struct {
	DocShortID uint64
	DocID      string
}

var _ Key = (*DocShortIDToDocIDAliasKey)(nil)

func NewDocShortIDToDocIDAliasKey(docShortID uint64, docID string) DocShortIDToDocIDAliasKey {
	return DocShortIDToDocIDAliasKey{
		DocShortID: docShortID,
		DocID:      docID,
	}
}

func (k DocShortIDToDocIDAliasKey) ToString() string {
	return string(k.Bytes())
}

func (k DocShortIDToDocIDAliasKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(DOC_SHORT_ID_TO_DOC_ID_ALIAS),
		EncodeDocShortID(k.DocShortID),
		stringSegment(k.DocID),
	)
}

func (k DocShortIDToDocIDAliasKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// BlockCIDToDocIDKey records one document owner for a document-owned block CID.
type BlockCIDToDocIDKey struct {
	BlockCID string
	DocID    string
}

var _ Key = (*BlockCIDToDocIDKey)(nil)

func NewBlockCIDToDocIDKey(blockCID string, docID string) BlockCIDToDocIDKey {
	return BlockCIDToDocIDKey{
		BlockCID: blockCID,
		DocID:    docID,
	}
}

func (k BlockCIDToDocIDKey) ToString() string {
	return string(k.Bytes())
}

func (k BlockCIDToDocIDKey) Bytes() []byte {
	return newDocIDSystemstoreKey(
		[]byte(BLOCK_CID_TO_DOC_ID),
		stringSegment(k.BlockCID),
		stringSegment(k.DocID),
	)
}

func (k BlockCIDToDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
