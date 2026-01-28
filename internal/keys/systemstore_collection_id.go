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

// CollectionID indexes collection short ids by the full id.
type CollectionID struct {
	CollectionID string
}

var _ Key = (*CollectionVersionKey)(nil)

func NewCollectionID(collectionID string) CollectionID {
	return CollectionID{
		CollectionID: collectionID,
	}
}

func NewCollectionIDFromString(keyString string) (CollectionID, error) {
	keyString = strings.TrimPrefix(keyString, COLLECTION_SHORT_ID+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 1 {
		return CollectionID{}, ErrInvalidKey
	}

	return CollectionID{
		CollectionID: elements[0],
	}, nil
}

func (k CollectionID) ToString() string {
	result := COLLECTION_SHORT_ID

	if k.CollectionID != "" {
		result = result + "/" + k.CollectionID
	}

	return result
}

func (k CollectionID) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionID) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
