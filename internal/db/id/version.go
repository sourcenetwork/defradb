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

// GetShortVersionID returns the local, shortened, internal, version id, which is used
// only in locations where using the full VID would be a waste of storage space.
func GetShortVersionID(
	ctx context.Context,
	collectionID, versionID string,
) (uint32, error) {
	cache := getVersionShortIDCache(ctx)
	shortID, ok := cache.Get(collectionID, versionID)
	if ok {
		return shortID, nil
	}

	key := keys.NewCollectionVersionKey(collectionID, versionID)

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

	cache.Set(collectionID, versionID, shortID)
	return shortID, nil
}

// SetShortVersionID sets and stores the short version id, if it does not already exist.
func SetShortVersionID(
	ctx context.Context,
	collectionID, versionID string,
) (uint32, error) {
	shortID, err := GetShortVersionID(ctx, collectionID, versionID)
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return 0, err
	}
	if shortID != 0 {
		return shortID, nil
	}

	colSeq, err := sequence.Get(ctx, keys.CollectionVersionIDSequenceKey{})
	if err != nil {
		return 0, err
	}

	sID, err := colSeq.Next(ctx)
	if err != nil {
		return 0, err
	}

	shortID = uint32(sID)

	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewCollectionVersionKey(collectionID, versionID)
	err = txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(shortID))))
	if err != nil {
		return 0, err
	}

	cache := getVersionShortIDCache(ctx)
	cache.Set(collectionID, versionID, shortID)

	return shortID, nil
}

type versionShortIDCacheKey struct{}

type versionShortIDCache map[string]map[string]uint32

func (c versionShortIDCache) Get(collectionID, versionID string) (uint32, bool) {
	if colCache, ok := c[collectionID]; ok {
		if shortID, ok := colCache[versionID]; ok {
			return shortID, true
		}
	}
	return 0, false
}

func (c versionShortIDCache) Set(collectionID, versionID string, shortID uint32) {
	if _, ok := c[collectionID]; !ok {
		c[collectionID] = make(map[string]uint32)
	}
	c[collectionID][versionID] = shortID
}

// InitVersionShortIDCache initializes the context with a none-nil version
// short-id cache.
//
// It is done to avoid an extra check to see if the cache exists or not when fetching
// it from the context.
func InitVersionShortIDCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, versionShortIDCacheKey{}, versionShortIDCache{})
}

// getVersionShortIDCache retrieves the version short-id cache from the given context.
func getVersionShortIDCache(ctx context.Context) versionShortIDCache {
	return ctx.Value(versionShortIDCacheKey{}).(versionShortIDCache) //nolint:forcetypeassert
}
