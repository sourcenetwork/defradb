// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/cache"
	"github.com/sourcenetwork/defradb/internal/db/lock"
)

type CollectionIndexKind int

const (
	CollectionVersionID CollectionIndexKind = 1
	CollectionID        CollectionIndexKind = 2
	CollectionName      CollectionIndexKind = 3
)

type CollectionIndex struct {
	Kind  CollectionIndexKind
	Value string
}

type CollectionRepository = cache.TxnRepository[CollectionIndex, client.CollectionVersion]

func NewColCache(lockSet *lock.LockSet, txnFreeDatastore corekv.Reader) *CollectionRepository {
	collectionStore := newCollectionStore(lockSet, txnFreeDatastore)

	collectionRepository := cache.NewTxnLayer(
		func() cache.Cache[CollectionIndex, client.CollectionVersion] {
			return newCollectionCache()
		},
		collectionStore,
	)

	collectionStore.collectionRepository = collectionRepository

	return collectionRepository
}
