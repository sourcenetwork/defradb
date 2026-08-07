// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/corekv/namespace"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/db/lock"
)

func TestMarkAsMerged_RolledBackTxnLeavesBlockUnmerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	// A fetched but unmerged block: stored, with its to-merge marker set.
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	block := putMarkedBlock(t, ctx, blockNS, []byte("rolled back"), newToMergeValue(time.Unix(1_700_000_000, 0)))

	txn := NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
	require.NoError(t, txn.Blockstore().MarkAsMerged(ctx, block.Cid()))
	txn.Discard()

	merged, err := BlockstoreFrom(rootstore, immutable.None[int]()).IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.False(t, merged)
}

func TestMarkAsMerged_CommittedTxnMarksBlockMerged(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	block := putMarkedBlock(t, ctx, blockNS, []byte("committed"), newToMergeValue(time.Unix(1_700_000_000, 0)))

	txn := NewTxnFrom(rootstore, lock.NewLockSet(), 1, false, immutable.None[int]())
	require.NoError(t, txn.Blockstore().MarkAsMerged(ctx, block.Cid()))
	require.NoError(t, txn.Commit())

	merged, err := BlockstoreFrom(rootstore, immutable.None[int]()).IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged)

	// The read populates the cache, so later reads skip the store.
	_, cached := globalMergedCache.Get(block.Cid().String())
	require.True(t, cached)
}
