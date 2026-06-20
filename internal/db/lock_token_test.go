// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
)

// lockToken must satisfy datastore.Txn so that read paths which do
// datastore.CtxMustGetTxn(ctx).Systemstore() keep working under it.
var _ datastore.Txn = (*lockToken)(nil)

func newTestLockToken(rootstore *memory.Datastore, lockSet *lock.LockSet, id uint64) *lockToken {
	return &lockToken{
		Multistore: datastore.NewMultistore(rootstore, lockSet, immutable.None[int]()),
		id:         id,
	}
}

// release must fire registered callbacks exactly once, even if called multiple times.
func TestLockToken_ReleaseIsIdempotent(t *testing.T) {
	calls := 0
	tok := &lockToken{id: 1}
	tok.OnDiscard(func() { calls++ })

	tok.release()
	tok.release()

	require.Equal(t, 1, calls, "release should fire registered callbacks exactly once")
}

// A collection write lock held under a lock token must block a competing read lock until the
// token is released, and releasing the token must free it. This mirrors how Truncate/RefreshView
// hold the collection write lock without an open store transaction.
func TestLockToken_ReleaseFreesCompetingRLock(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	lockSet := lock.NewLockSet()

	const shortID = uint32(7)
	writer := newTestLockToken(rootstore, lockSet, 1)
	reader := newTestLockToken(rootstore, lockSet, 2)

	lockSet.CollectionLock(writer, shortID)

	acquired := make(chan struct{})
	go func() {
		lockSet.CollectionRLock(reader, shortID)
		close(acquired)
	}()

	select {
	case <-acquired:
		t.Fatal("CollectionRLock acquired while a write lock was held; expected it to block")
	case <-time.After(200 * time.Millisecond):
		// Expected: the read lock is blocked behind the write lock.
	}

	writer.release()

	select {
	case <-acquired:
		// Expected: releasing the token frees the write lock so the read lock can be acquired.
	case <-time.After(2 * time.Second):
		t.Fatal("CollectionRLock did not acquire after the write-lock token was released")
	}
}
