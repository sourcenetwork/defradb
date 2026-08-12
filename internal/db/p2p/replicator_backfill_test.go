// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// multistoreDB satisfies DB for the backfill path, which reaches no other method.
type multistoreDB struct {
	DB
	stores *datastore.Multistore
}

func (d multistoreDB) Multistore() *datastore.Multistore { return d.stores }

// backfillCollection carries only the identity the backfill reads off a collection.
type backfillCollection struct {
	client.Collection
	collectionID string
}

func (c backfillCollection) CollectionID() string { return c.collectionID }

func (c backfillCollection) Version() client.CollectionVersion {
	return client.CollectionVersion{CollectionID: c.collectionID}
}

// A purge that leaves a document's primary key behind, or any other source of one, must not
// stop the backfill. The iterator walks the primary prefix in key order, so returning on the
// first key that will not resolve skips every document sorting after it.
func TestPushHeadsForAllDocs_SkipsPrimaryKeysWithNoDocument(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	stores := datastore.NewMultistore(rootstore, lock.NewLockSet(), immutable.None[int]())

	const collectionID = "bafkreicollection"
	const shortID = 1
	require.NoError(t, stores.Systemstore().Set(ctx,
		keys.NewCollectionID(collectionID).Bytes(), []byte(strconv.Itoa(shortID))))

	// Three primary keys, none of which has a docID mapping behind it. Written through the
	// same unsafe view the backfill iterates, since the locked wrapper wants a txn in context.
	unsafe, ok := stores.Datastore().(interface{ Unsafe() corekv.ReaderWriter })
	require.True(t, ok, "the backfill reaches the datastore through Unsafe")
	raw := unsafe.Unsafe()
	for _, docShortID := range []uint64{1, 2, 3} {
		key := keys.PrimaryDataStoreKey{CollectionShortID: shortID, DocShortID: docShortID}
		require.NoError(t, raw.Set(ctx, key.Bytes(), []byte{}))
	}

	p := &P2P{db: multistoreDB{stores: stores}}
	col := backfillCollection{collectionID: collectionID}

	require.NoError(t, p.pushHeadsForAllDocs(ctx, col, "peer"),
		"a primary key with no document must be skipped, not returned as an error")
}
