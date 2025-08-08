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
)

// CollectionVersionKey indexes collection version ids by their collection id.
//
// The index is the key, there are no values stored against the key.
type CollectionVersionKey struct {
	CollectionID string
	VersionID    string
}

var _ Key = (*CollectionVersionKey)(nil)

func NewCollectionVersionKey(collectionID string, versionID string) CollectionVersionKey {
	return CollectionVersionKey{
		CollectionID: collectionID,
		VersionID:    versionID,
	}
}

func NewCollectionVersionKeyFromString(keyString string) (CollectionVersionKey, error) {
	keyString = strings.TrimPrefix(keyString, COLLECTION_VERSION+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return CollectionVersionKey{}, ErrInvalidKey
	}

	return CollectionVersionKey{
		CollectionID: elements[0],
		VersionID:    elements[1],
	}, nil
}

func (k CollectionVersionKey) ToString() string {
	result := COLLECTION_VERSION

	if k.CollectionID != "" {
		result = result + "/" + k.CollectionID
	}

	if k.VersionID != "" {
		result = result + "/" + k.VersionID
	}

	return result
}

func (k CollectionVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
