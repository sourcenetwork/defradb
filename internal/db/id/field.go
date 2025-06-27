// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"
	"errors"
	"strconv"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// GetShortFieldID returns the local, shortened, internal, field id, which is used
// only in locations where using the full CID would be a waste of storage space.
func GetShortFieldID(
	ctx context.Context,
	collectionShortID uint32,
	fieldID string,
) (uint32, error) {
	// This concatenation is temporary, soon we can just use the field CID
	uniqueKey := strconv.Itoa(int(collectionShortID)) + ":" + fieldID

	cache := getFieldShortIDCache(ctx)
	shortID, ok := cache[uniqueKey]
	if ok {
		return shortID, nil
	}

	// If we miss the cache, load the entire collection's worth - it is almost always
	// going to be more efficient than loading the field short-ids one by one, and we'll
	// usually want most of them.

	key := keys.NewFieldIDPrefix(collectionShortID)
	txn := datastore.CtxMustGetTxn(ctx)
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{Prefix: key.Bytes()})
	if err != nil {
		return 0, err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return 0, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key, err := keys.NewFieldIDFromString(string(iter.Key()))
		if err != nil {
			return 0, errors.Join(err, iter.Close())
		}

		value, err := iter.Value()
		if err != nil {
			return 0, errors.Join(err, iter.Close())
		}

		v, err := strconv.ParseUint(string(value), 10, 0)
		if err != nil {
			return 0, err
		}
		sID := uint32(v)

		// This concatenation is temporary, soon we can just use the field CID
		uniqueKey := strconv.Itoa(int(collectionShortID)) + ":" + key.FieldID
		cache[uniqueKey] = sID

		if key.FieldID == fieldID {
			shortID = sID
		}
	}

	return shortID, iter.Close()
}

// SetShortFieldID sets and stores the short field id, if it does not already exist.
func SetShortFieldID(
	ctx context.Context,
	collectionShortID uint32,
	fieldID string,
) error {
	// This concatenation is temporary, soon we can just use the field CID
	uniqueKey := strconv.Itoa(int(collectionShortID)) + ":" + fieldID

	cache := getFieldShortIDCache(ctx)
	_, ok := cache[uniqueKey]
	if ok {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewFieldID(collectionShortID, fieldID)

	hasShortID, err := txn.Systemstore().Has(ctx, key.Bytes())
	if err != nil {
		return err
	}
	if hasShortID {
		return nil
	}

	fieldSeq, err := sequence.Get(ctx, keys.NewFieldIDSequenceKey(collectionShortID))
	if err != nil {
		return err
	}

	sID, err := fieldSeq.Next(ctx)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(sID))))
	if err != nil {
		return err
	}

	cache[uniqueKey] = uint32(sID)

	return nil
}

// SetShortFieldID sets and stores the short field ids, if they do not already exist.
func SetShortFieldIDs(ctx context.Context, collection client.CollectionVersion) error {
	collectionShortID, err := GetShortCollectionID(ctx, collection.CollectionID)
	if err != nil {
		return err
	}

	for _, field := range collection.Fields {
		err := SetShortFieldID(ctx, collectionShortID, field.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

type fieldShortIDCacheKey struct{}

// fieldShortIDCache contains field short-ids by a concatenation of [CollectionID][FieldName].
//
// In the near future the key will be replaced by the field cid.
type fieldShortIDCache map[string]uint32

// InitCollectionShortIDCache initialializes the context with a none-nil collection
// short-id cache.
//
// It is done to avoid an extra check to see if the cache exists or not when fetching
// it from the context.
func InitFieldShortIDCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, fieldShortIDCacheKey{}, fieldShortIDCache{})
}

// getCollectionShortIDCache retrieves the collection short-id cache from the given context.
func getFieldShortIDCache(ctx context.Context) fieldShortIDCache {
	return ctx.Value(fieldShortIDCacheKey{}).(fieldShortIDCache) //nolint:forcetypeassert
}
