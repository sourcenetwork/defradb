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
	"bytes"
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
	// CollectionShortID is the id of the Collection that this item belongs to.
	CollectionShortID uint32

	// ItemID is the unique (to this CollectionRootID) ID of the View item.
	//
	// For now this is essentially just the index of the item in the result-set, however
	// that is likely to change in the near future.
	ItemID uint
}

var _ Key = (*ViewCacheKey)(nil)
var _ CollectionedKey = ViewCacheKey{}

func NewViewCacheColPrefix(collectionShortID uint32) ViewCacheKey {
	return ViewCacheKey{
		CollectionShortID: collectionShortID,
	}
}

func NewViewCacheKey(collectionShortID uint32, itemID uint) ViewCacheKey {
	return ViewCacheKey{
		CollectionShortID: collectionShortID,
		ItemID:            itemID,
	}
}

func NewViewCacheKeyFromRaw(raw []byte) (ViewCacheKey, error) {
	if len(raw) == 0 {
		return ViewCacheKey{}, nil
	}

	raw, _ = bytes.CutPrefix(raw, []byte(COLLECTION_VIEW_ITEMS+"/"))

	components := bytes.Split(raw, []byte("/"))
	if len(components) > 2 {
		return ViewCacheKey{}, ErrInvalidKey
	}

	_, collectionShortID, err := encoding.DecodeUvarintAscending(components[0])
	if err != nil {
		return ViewCacheKey{}, err
	}

	var itemID uint
	if len(components) == 2 {
		_, r, err := encoding.DecodeUvarintAscending(components[1])
		if err != nil {
			return ViewCacheKey{}, err
		}
		itemID = uint(r)
	}

	return ViewCacheKey{
		CollectionShortID: uint32(collectionShortID),
		ItemID:            itemID,
	}, nil
}

func (k ViewCacheKey) ToString() string {
	return string(k.Bytes())
}

func (k ViewCacheKey) Bytes() []byte {
	result := []byte(COLLECTION_VIEW_ITEMS)

	if k.CollectionShortID != 0 {
		result = append(result, '/')
		result = encoding.EncodeUvarintAscending(result, uint64(k.CollectionShortID))
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

	if k.CollectionShortID != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionShortID))
	}
	if k.ItemID != 0 {
		result = result + "/" + strconv.Itoa(int(k.ItemID))
	}

	return result
}

func (k ViewCacheKey) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}
