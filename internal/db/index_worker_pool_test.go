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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// TestBuildPool_CapsLiveWorkers checks the pool never runs more than indexBuildConcurrency workers at
// once, even with more pending indexes than the cap. Each build parks at the gate so many are pending
// together; a peak counter sampled at the gate must never exceed the cap.
func TestBuildPool_CapsLiveWorkers(t *testing.T) {
	const cap = 3
	setForTest(t, &indexBuildConcurrency, cap)
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	// Gate: park every build in the gate at once so concurrency is observable, tracking the peak
	// number of builds parked together. release lets them all proceed once the peak is sampled.
	origGate := IndexBuildGate
	t.Cleanup(func() { IndexBuildGate = origGate })
	release := make(chan struct{})
	var releaseOnce sync.Once
	var inGate atomic.Int32
	var peak atomic.Int32
	IndexBuildGate = func(gateCtx context.Context, _ string, _ uint32) {
		select {
		case <-release:
			return
		default:
		}
		cur := inGate.Add(1)
		for {
			p := peak.Load()
			if cur <= p || peak.CompareAndSwap(p, cur) {
				break
			}
		}
		select {
		case <-release:
		case <-gateCtx.Done():
		}
		inGate.Add(-1)
	}

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { a: Int, b: Int, c: Int, d: Int, e: Int, f: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 20
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx,
			fmt.Appendf(nil, `{"a":%d,"b":%d,"c":%d,"d":%d,"e":%d,"f":%d}`, i, i, i, i, i, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	fields := []string{"a", "b", "c", "d", "e", "f"} // 6 indexes, twice the cap
	ids := make([]uint32, 0, len(fields))
	for _, f := range fields {
		desc, err := col.NewIndex(ctx, client.NewIndexRequest{
			Fields: []client.IndexedFieldDescription{{Name: f}},
		})
		require.NoError(t, err)
		ids = append(ids, desc.ID)
	}

	// Let the pool fill to its peak, then release. Sampling after a short settle is enough: workers
	// spawn eagerly on dispatch, so the peak is reached well before this.
	require.Eventually(t, func() bool {
		return inGate.Load() >= int32(cap)
	}, 20*time.Second, 5*time.Millisecond, "pool never reached its cap")
	assert.LessOrEqual(t, peak.Load(), int32(cap), "live workers must never exceed the pool cap")
	releaseOnce.Do(func() { close(release) })

	collectionID := col.Version().CollectionID
	for _, id := range ids {
		waitForIndexReady(t, ctx, db, collectionID, id)
	}
}

// TestBuildPool_QueueFull_AllIndexesStillBuild forces the task queue to overflow (size 1 with more
// pending indexes than that) and checks the drop-and-re-derive path loses no work: every index still
// converges to ready. The queue-full branch drops the task without notifying and relies on a busy
// worker's completion re-drive to pick the index up later.
func TestBuildPool_QueueFull_AllIndexesStillBuild(t *testing.T) {
	setForTest(t, &indexTaskQueueSize, 1)
	setForTest(t, &indexBuildConcurrency, 1)
	setForTest(t, &indexBackfillBatchSize, 5)
	setForTest(t, &indexBuildRetryDelay, 10*time.Millisecond)

	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { a: Int, b: Int, c: Int, d: Int, e: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	const docCount = 15
	for i := range docCount {
		doc, err := client.NewDocFromJSON(ctx,
			fmt.Appendf(nil, `{"a":%d,"b":%d,"c":%d,"d":%d,"e":%d}`, i, i, i, i, i), col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))
	}

	fields := []string{"a", "b", "c", "d", "e"} // 5 indexes, queue holds 1
	ids := make([]uint32, 0, len(fields))
	for _, f := range fields {
		desc, err := col.NewIndex(ctx, client.NewIndexRequest{
			Fields: []client.IndexedFieldDescription{{Name: f}},
		})
		require.NoError(t, err)
		ids = append(ids, desc.ID)
	}

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	for _, id := range ids {
		waitForIndexReady(t, ctx, db, collectionID, id)
		assert.Equal(t, docCount, countIndexEntries(t, ctx, db, shortID, id),
			"index %d must have one entry per doc despite queue overflow", id)
	}
}
