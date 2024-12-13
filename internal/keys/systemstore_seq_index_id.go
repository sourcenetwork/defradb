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
	"strconv"

	ds "github.com/ipfs/go-datastore"
)

// IndexIDSequenceKey is used to key the sequence used to generate index ids.
//
// The sequence is specific to each collection version.
type IndexIDSequenceKey struct {
	CollectionID uint32
}

var _ Key = (*IndexIDSequenceKey)(nil)

func NewIndexIDSequenceKey(collectionID uint32) IndexIDSequenceKey {
	return IndexIDSequenceKey{CollectionID: collectionID}
}

func (k IndexIDSequenceKey) ToString() string {
	return INDEX_ID_SEQ + "/" + strconv.Itoa(int(k.CollectionID))
}

func (k IndexIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
