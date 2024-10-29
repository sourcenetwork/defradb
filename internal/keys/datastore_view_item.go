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

	"github.com/sourcenetwork/defradb/internal/encoding"
)

// ViewCacheKey is a trimmed down [DataStoreKey] used for caching the results
// of View items.
//
// It is stored in the format `/collection/vi/[CollectionRootID]/[ItemID]`. It points to the
// full serialized View item.
type ViewCacheKey struct {
	// CollectionRootID is the Root of the Collection that this item belongs to.
	CollectionRootID uint32

	// ItemID is the unique (to this CollectionRootID) ID of the View item.
	//
	// For now this is essentially just the index of the item in the result-set, however
	// that is likely to change in the near future.
	ItemID uint
}

var _ Key = (*ViewCacheKey)(nil)

func NewViewCacheColPrefix(rootID uint32) ViewCacheKey {
	return ViewCacheKey{
		CollectionRootID: rootID,
	}
}

func NewViewCacheKey(rootID uint32, itemID uint) ViewCacheKey {
	return ViewCacheKey{
		CollectionRootID: rootID,
		ItemID:           itemID,
	}
}

func (k ViewCacheKey) ToString() string {
	return string(k.Bytes())
}

func (k ViewCacheKey) Bytes() []byte {
	result := []byte(COLLECTION_VIEW_ITEMS)

	if k.CollectionRootID != 0 {
		result = append(result, '/')
		result = encoding.EncodeUvarintAscending(result, uint64(k.CollectionRootID))
	}

	if k.ItemID != 0 {
		result = append(result, '/')
		result = encoding.EncodeUvarintAscending(result, uint64(k.ItemID))
	}

	return result
}

func (k ViewCacheKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k ViewCacheKey) PrettyPrint() string {
	result := COLLECTION_VIEW_ITEMS

	if k.CollectionRootID != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionRootID))
	}
	if k.ItemID != 0 {
		result = result + "/" + strconv.Itoa(int(k.ItemID))
	}

	return result
}
