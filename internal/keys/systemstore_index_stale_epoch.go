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
	"fmt"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// IndexStaleEpochKey marks an index that has superseded epochs left by a rebuild. A rebuild writes
// it when it advances the epoch sequence; the worker's stale-epoch sweep clears it once those epochs
// are collected. Without the marker the sweep would have to scan every index on every drain, so it
// is what lets the sweep run only for indexes that actually have stale entries.
//
// CollectionShortID is keyed by the stable collection ID, so the marker survives version switches.
type IndexStaleEpochKey struct {
	CollectionShortID uint32
	IndexID           uint32
}

var _ Key = (*IndexStaleEpochKey)(nil)

func NewIndexStaleEpochKey(collectionShortID, indexID uint32) IndexStaleEpochKey {
	return IndexStaleEpochKey{CollectionShortID: collectionShortID, IndexID: indexID}
}

// IndexStaleEpochPrefix is the prefix over every stale-epoch marker, used to scan them all.
func IndexStaleEpochPrefix() string {
	return INDEX_STALE_EPOCH
}

// NewIndexStaleEpochKeyFromString parses a marker key produced by ToString.
func NewIndexStaleEpochKeyFromString(key string) (IndexStaleEpochKey, error) {
	trimmed := strings.TrimPrefix(key, INDEX_STALE_EPOCH+"/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return IndexStaleEpochKey{}, fmt.Errorf("invalid stale-epoch key: %q", key)
	}
	collectionShortID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return IndexStaleEpochKey{}, fmt.Errorf("invalid stale-epoch key %q: %w", key, err)
	}
	indexID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return IndexStaleEpochKey{}, fmt.Errorf("invalid stale-epoch key %q: %w", key, err)
	}
	return IndexStaleEpochKey{CollectionShortID: uint32(collectionShortID), IndexID: uint32(indexID)}, nil
}

func (k IndexStaleEpochKey) ToString() string {
	return INDEX_STALE_EPOCH + "/" +
		strconv.FormatUint(uint64(k.CollectionShortID), 10) + "/" +
		strconv.FormatUint(uint64(k.IndexID), 10)
}

func (k IndexStaleEpochKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexStaleEpochKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
