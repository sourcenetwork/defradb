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
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// IndexStateKey is used to key the mutable runtime state of an index, stored in the
// systemstore separately from the immutable index description.
//
// CollectionID is the stable collection ID (not a version ID), so state survives
// collection version switches.
type IndexStateKey struct {
	CollectionID string
	IndexID      uint32
}

var _ Key = (*IndexStateKey)(nil)

// NewIndexStateKey returns an IndexStateKey for the given collection and index.
func NewIndexStateKey(collectionID string, indexID uint32) IndexStateKey {
	return IndexStateKey{
		CollectionID: collectionID,
		IndexID:      indexID,
	}
}

// NewIndexStateKeyFromString parses a full key string back into an IndexStateKey.
//
// The expected format is: INDEX_STATE + "/" + collectionID + "/" + indexID
func NewIndexStateKeyFromString(s string) (IndexStateKey, error) {
	prefix := INDEX_STATE + "/"
	if !strings.HasPrefix(s, prefix) {
		return IndexStateKey{}, ErrInvalidKey
	}

	remainder := strings.TrimPrefix(s, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return IndexStateKey{}, ErrInvalidKey
	}

	id, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return IndexStateKey{}, ErrInvalidKey
	}

	return IndexStateKey{
		CollectionID: parts[0],
		IndexID:      uint32(id),
	}, nil
}

// NewIndexStateKeyPrefix returns the byte prefix used to iterate over all index state entries
// across all collections. This is useful for startup recovery scans.
func NewIndexStateKeyPrefix() []byte {
	return []byte(INDEX_STATE + "/")
}

// NewIndexStateCollectionPrefix returns the byte prefix used to iterate over all index state
// entries for a specific collection.
func NewIndexStateCollectionPrefix(collectionID string) []byte {
	return []byte(INDEX_STATE + "/" + collectionID + "/")
}

func (k IndexStateKey) ToString() string {
	return INDEX_STATE + "/" + k.CollectionID + "/" + strconv.FormatUint(uint64(k.IndexID), 10)
}

func (k IndexStateKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexStateKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
