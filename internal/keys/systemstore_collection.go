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
	"fmt"

	ds "github.com/ipfs/go-datastore"
)

// CollectionKey points to the json serialized description of the
// the collection of the given ID.
type CollectionKey struct {
	CollectionID string
}

var _ Key = (*CollectionKey)(nil)

// Returns a formatted collection key for the system data store.
// It assumes the id of the collection is non-empty.
func NewCollectionKey(id string) CollectionKey {
	return CollectionKey{CollectionID: id}
}

func (k CollectionKey) ToString() string {
	return fmt.Sprintf("%s/%s", COLLECTION_ID, k.CollectionID)
}

func (k CollectionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
