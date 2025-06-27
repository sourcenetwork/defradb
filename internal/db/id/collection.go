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
	"strconv"

	"github.com/sourcenetwork/defradb/datastore"
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
) error {
	cache := getCollectionShortIDCache(ctx)
	_, ok := cache[collectionID]
	if ok {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewCollectionID(collectionID)

	hasShortID, err := txn.Systemstore().Has(ctx, key.Bytes())
	if err != nil {
		return err
	}
	if hasShortID {
		return nil
	}

	colSeq, err := sequence.Get(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	sID, err := colSeq.Next(ctx)
	if err != nil {
		return err
	}
	shortID := uint32(sID)

	err = txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(shortID))))
	if err != nil {
		return err
	}

	cache[collectionID] = shortID

	return nil
}

type collectionShortIDCacheKey struct{}

type collectionShortIDCache map[string]uint32

// InitCollectionShortIDCache initialializes the context with a none-nil collection
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
