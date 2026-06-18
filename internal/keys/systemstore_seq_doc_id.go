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

// DocIDSequenceKey is used to key the per-collection short document ID sequence.
type DocIDSequenceKey struct {
	CollectionShortID uint32
}

var _ Key = (*DocIDSequenceKey)(nil)

func NewDocIDSequenceKey(collectionShortID uint32) DocIDSequenceKey {
	return DocIDSequenceKey{CollectionShortID: collectionShortID}
}

func (k DocIDSequenceKey) ToString() string {
	return DOC_ID_SEQ + "/" + strconv.Itoa(int(k.CollectionShortID))
}

func (k DocIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k DocIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
