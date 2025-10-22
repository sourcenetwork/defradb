// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// GetShortCollectionID returns the local, shortened, internal, collection id, which is used
// only in locations where using the full CID would be a waste of storage space.
func GetShortCollectionID(
	ctx context.Context,
	collectionID string,
) (uint32, error) {
	cache := getCollectionShortIDCache(ctx)
	shortID, ok := cache[collectionID]
	if ok {
		return shortID, nil
	}

	key := keys.NewCollectionID(collectionID)

	txn := datastore.CtxMustGetTxn(ctx)
	valueBytes, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseUint(string(valueBytes), 10, 0)
	if err != nil {
		return 0, err
	}
	shortID = uint32(v)

	cache[collectionID] = shortID
	return shortID, nil
}

// SetShortCollectionID sets and stores the short collection id, if it does not already exist.
func SetShortCollectionID(
	ctx context.Context,
	collectionID string,
) (uint32, error) {
	shortID, err := GetShortCollectionID(ctx, collectionID)
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return 0, err
	}
	if shortID != 0 {
		return shortID, nil
	}

	colSeq, err := sequence.Get(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return 0, err
	}

	sID, err := colSeq.Next(ctx)
	if err != nil {
		return 0, err
	}
	shortID = uint32(sID)

	txn := datastore.CtxMustGetTxn(ctx)
	err = txn.Systemstore().Set(ctx, keys.NewCollectionID(collectionID).Bytes(), []byte(strconv.Itoa(int(shortID))))
	if err != nil {
		return 0, err
	}

	cache := getCollectionShortIDCache(ctx)
	cache[collectionID] = shortID

	return shortID, nil
}

type collectionShortIDCacheKey struct{}

type collectionShortIDCache map[string]uint32

// InitCollectionShortIDCache initializes the context with a none-nil collection
// short-id cache.
//
// It is done to avoid an extra check to see if the cache exists or not when fetching
// it from the context.
func InitCollectionShortIDCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, collectionShortIDCacheKey{}, collectionShortIDCache{})
}

// getCollectionShortIDCache retrieves the collection short-id cache from the given context.
func getCollectionShortIDCache(ctx context.Context) collectionShortIDCache {
	return ctx.Value(collectionShortIDCacheKey{}).(collectionShortIDCache) //nolint:forcetypeassert
}
