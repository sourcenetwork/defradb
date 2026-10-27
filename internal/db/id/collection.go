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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// GetUncachedCollectionShortID returns the collection short ID used in storage keys.
//
// GetCollectionShortID should be preferred over this method because it utilizes the cache.
func GetUncachedCollectionShortID(
	ctx context.Context,
	collectionID string,
	systemStore corekv.ReaderWriter,
) (uint32, error) {
	key := keys.NewCollectionID(collectionID)
	valueBytes, err := systemStore.Get(ctx, key.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			err = NewErrGetCollectionShortID(client.ErrCollectionNotFound, collectionID)
		}
		return 0, err
	}
	v, err := strconv.ParseUint(string(valueBytes), 10, 0)
	if err != nil {
		return 0, NewErrParseCollectionShortID(err, collectionID)
	}
	return uint32(v), nil
}

// GetCollectionShortID returns the collection short ID used in storage keys.
func GetCollectionShortID(
	ctx context.Context,
	collectionID string,
) (uint32, error) {
	cache, ok := ctx.Value(collectionShortIDCacheKey{}).(collectionShortIDCache)
	if ok {
		collectionShortID, ok := cache[collectionID]
		if ok {
			return collectionShortID, nil
		}
	}
	txn := datastore.CtxMustGetTxn(ctx)
	collectionShortID, err := GetUncachedCollectionShortID(ctx, collectionID, txn.Systemstore())
	if err != nil {
		return 0, err
	}
	if ok {
		cache[collectionID] = collectionShortID
	}
	return collectionShortID, nil
}

// SetCollectionShortID stores the collection short ID, if it does not already exist.
func SetCollectionShortID(
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
		return NewErrCheckCollectionShortID(err, collectionID)
	}
	if hasShortID {
		return nil
	}

	colSeq, err := sequence.Get(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return NewErrGetCollectionIDSequence(err, collectionID)
	}

	sID, err := colSeq.Next(ctx)
	if err != nil {
		return NewErrNextCollectionIDSeq(err, collectionID)
	}
	collectionShortID := uint32(sID)

	err = txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(collectionShortID))))
	if err != nil {
		return NewErrStoreCollectionShortID(err, collectionID)
	}

	cache[collectionID] = collectionShortID

	return nil
}

func DeleteCollectionShortID(
	ctx context.Context,
	collectionID string,
) error {
	cache := getCollectionShortIDCache(ctx)
	delete(cache, collectionID)

	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewCollectionID(collectionID)

	err := txn.Systemstore().Delete(ctx, key.Bytes())
	if err != nil {
		return NewErrDeleteCollectionShortID(err, collectionID)
	}
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
