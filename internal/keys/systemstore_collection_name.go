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

// CollectionNameKey points to the ID of the collection of the given
// name.
type CollectionNameKey struct {
	Name string
}

var _ Key = (*CollectionNameKey)(nil)

func NewCollectionNameKey(name string) CollectionNameKey {
	return CollectionNameKey{Name: name}
}

func (k CollectionNameKey) ToString() string {
	return fmt.Sprintf("%s/%s", COLLECTION_NAME, k.Name)
}

func (k CollectionNameKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionNameKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
