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

// IndexMutationCounterKey holds a single monotonic counter per collection, bumped whenever something
// changes what a write must maintain: an index created, dropped, rebuilt (epoch advanced), or a build
// finishing. A write reads this counter to decide whether its cached index set is still current
// without re-resolving the whole set, and reading it also puts it in the write's read-set so a
// concurrent index change conflicts the write and it retries.
//
// CollectionShortID is keyed by the stable collection ID, so the counter survives version switches.
type IndexMutationCounterKey struct {
	CollectionShortID uint32
}

var _ Key = (*IndexMutationCounterKey)(nil)

func NewIndexMutationCounterKey(collectionShortID uint32) IndexMutationCounterKey {
	return IndexMutationCounterKey{CollectionShortID: collectionShortID}
}

func (k IndexMutationCounterKey) ToString() string {
	return INDEX_MUTATION_COUNTER + "/" +
		strconv.FormatUint(uint64(k.CollectionShortID), 10)
}

func (k IndexMutationCounterKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexMutationCounterKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
