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
	"strconv"

	ds "github.com/ipfs/go-datastore"
)

// Doc ID mapping keys bridge two lookup shapes:
// collection-scoped keys preserve the short ID <-> public DocID relation, while
// node-scoped keys handle paths that only know a public DocID. Genesis field
// indexes are stored both ways so verification can resolve field blocks quickly
// and document cleanup can remove those reverse indexes without scanning.
const (
	SHORT_ID_TO_DOC_ID      = "s"
	DOC_ID_TO_SHORT_ID      = "p"
	NODE_DOC_ID_INDEX       = "n"
	GENESIS_FIELD_TO_DOC_ID = "f"
	DOC_ID_TO_GENESIS_FIELD = "pf"
)

type systemstoreDocIDKey struct {
	segments []string
}

func newSystemstoreDocIDKey(segments ...string) systemstoreDocIDKey {
	return systemstoreDocIDKey{segments: segments}
}

func (k systemstoreDocIDKey) ToString() string {
	result := DOC_ID_INDEX
	for _, segment := range k.segments {
		if segment != "" {
			result += "/" + segment
		}
	}
	return result
}

func (k systemstoreDocIDKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k systemstoreDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// ShortIDToDocIDKey maps a short doc ID to its public doc ID.
type ShortIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*ShortIDToDocIDKey)(nil)

func NewShortIDToDocIDKey(collectionShortID uint32, shortDocID string) ShortIDToDocIDKey {
	return ShortIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			strconv.Itoa(int(collectionShortID)),
			SHORT_ID_TO_DOC_ID,
			shortDocID,
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
			strconv.Itoa(int(collectionShortID)),
			DOC_ID_TO_SHORT_ID,
			docID,
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
			NODE_DOC_ID_INDEX,
			DOC_ID_TO_SHORT_ID,
			docID,
		),
	}
}

// NodeShortIDToDocIDKey maps this node's local short doc ID to its public doc ID.
type NodeShortIDToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*NodeShortIDToDocIDKey)(nil)

func NewNodeShortIDToDocIDKey(shortDocID string) NodeShortIDToDocIDKey {
	return NodeShortIDToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			NODE_DOC_ID_INDEX,
			SHORT_ID_TO_DOC_ID,
			shortDocID,
		),
	}
}

// GenesisFieldToDocIDKey maps a genesis field block CID to one public doc ID that links to it.
type GenesisFieldToDocIDKey struct {
	systemstoreDocIDKey
}

var _ Key = (*GenesisFieldToDocIDKey)(nil)

func NewGenesisFieldToDocIDKey(collectionShortID uint32, fieldCID string, docID string) GenesisFieldToDocIDKey {
	return GenesisFieldToDocIDKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			strconv.Itoa(int(collectionShortID)),
			GENESIS_FIELD_TO_DOC_ID,
			fieldCID,
			docID,
		),
	}
}

// DocIDToGenesisFieldKey maps a public doc ID to one of its genesis field block CIDs.
type DocIDToGenesisFieldKey struct {
	systemstoreDocIDKey
}

var _ Key = (*DocIDToGenesisFieldKey)(nil)

func NewDocIDToGenesisFieldKey(collectionShortID uint32, docID string, fieldCID string) DocIDToGenesisFieldKey {
	return DocIDToGenesisFieldKey{
		systemstoreDocIDKey: newSystemstoreDocIDKey(
			strconv.Itoa(int(collectionShortID)),
			DOC_ID_TO_GENESIS_FIELD,
			docID,
			fieldCID,
		),
	}
}
