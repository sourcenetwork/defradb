// Copyright 2024 Democratized Data Foundation
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
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"
)

// CollectionIndexKey to a stored description of an index
type CollectionIndexKey struct {
	// CollectionID is the root identifier id of the collection that the index is on
	CollectionID immutable.Option[string]
	// IndexName is the name of the index
	IndexName string
}

var _ Key = (*CollectionIndexKey)(nil)

// NewCollectionIndexKey creates a new CollectionIndexKey from a collection root and index name.
func NewCollectionIndexKey(collectionID immutable.Option[string], indexName string) CollectionIndexKey {
	return CollectionIndexKey{CollectionID: collectionID, IndexName: indexName}
}

// NewCollectionIndexKeyFromString creates a new CollectionIndexKey from a string.
// It expects the input string is in the following format:
//
// /collection/index/[CollectionID]/[IndexName]
//
// Where [IndexName] might be omitted. Anything else will return an error.
func NewCollectionIndexKeyFromString(key string) (CollectionIndexKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) < 4 || len(keyArr) > 5 || keyArr[1] != COLLECTION || keyArr[2] != "index" {
		return CollectionIndexKey{}, ErrInvalidKey
	}

	result := CollectionIndexKey{CollectionID: immutable.Some(keyArr[3])}
	if len(keyArr) == 5 {
		result.IndexName = keyArr[4]
	}
	return result, nil
}

// ToString returns the string representation of the key
// It is in the following format:
// /collection/index/[CollectionID]/[IndexName]
// if [CollectionID] is empty, the rest is ignored
func (k CollectionIndexKey) ToString() string {
	result := COLLECTION_INDEX

	if k.CollectionID.HasValue() {
		result = result + "/" + k.CollectionID.Value()
		if k.IndexName != "" {
			result = result + "/" + k.IndexName
		}
	}

	return result
}

// Bytes returns the byte representation of the key
func (k CollectionIndexKey) Bytes() []byte {
	return []byte(k.ToString())
}

// ToDS returns the datastore key
func (k CollectionIndexKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
