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

// IndexEpochSequenceKey keys the sequence that allocates index entry epochs.
//
// The sequence is specific to a single index. Each rebuild draws the next value as its
// building epoch, so a building epoch never reuses the active epoch or one left behind by
// an interrupted rebuild whose entries have not yet been collected.
//
// CollectionID is the stable collection ID, so the sequence survives version switches.
type IndexEpochSequenceKey struct {
	CollectionID string
	IndexID      uint32
}

var _ Key = (*IndexEpochSequenceKey)(nil)

func NewIndexEpochSequenceKey(collectionID string, indexID uint32) IndexEpochSequenceKey {
	return IndexEpochSequenceKey{CollectionID: collectionID, IndexID: indexID}
}

func (k IndexEpochSequenceKey) ToString() string {
	return INDEX_EPOCH_SEQ + "/" + k.CollectionID + "/" + strconv.FormatUint(uint64(k.IndexID), 10)
}

func (k IndexEpochSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexEpochSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
