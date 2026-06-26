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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// TestIndexWorker_DrainNoRecords_NoOp checks that draining a collection with no pending build or
// drop records does nothing and leaves no state.
func TestIndexWorker_DrainNoRecords_NoOp(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	addUserDoc(t, ctx, col, "Alice")

	// No index, so no records: the drain must be a no-op and return promptly.
	db.indexBuildWorker.drainSync(ctx)

	indexes, err := col.ListIndexes(ctx)
	require.NoError(t, err)
	assert.Empty(t, indexes)
}

// TestIndexWorker_DrainBuildingRecord_Completes checks that a drain over a committed building
// record backfills the index and clears the record.
func TestIndexWorker_DrainBuildingRecord_Completes(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	names := []string{"Alice", "Bob", "Carol"}
	for _, name := range names {
		addUserDoc(t, ctx, col, name)
	}

	// NewIndex stages the building record; the worker (driven here by drainSync) builds it.
	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	state := readIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
	require.True(t, state.isBuilding(), "index must be building before the drain")

	db.indexBuildWorker.drainSync(ctx)

	requireNoIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
	for _, name := range names {
		require.Len(t, queryUserByName(t, db, ctx, name), 1, "name %q must be queryable", name)
	}
}

// TestIndexWorker_InFlightGuard_PreventsDoubleBuild checks that while a build for an index is in
// flight, a concurrent drain does not start a second build for the same index. The gate blocks the
// first build mid-way so a second drain runs against the still-present record.
func TestIndexWorker_InFlightGuard_PreventsDoubleBuild(t *testing.T) {
	origGate := IndexBuildGate
	defer func() { IndexBuildGate = origGate }()

	var starts atomic.Int32
	release := make(chan struct{})
	entered := make(chan struct{}, 1)
	IndexBuildGate = func(_ context.Context, _ string, _ uint32) {
		starts.Add(1)
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
	}

	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	addUserDoc(t, ctx, col, "Alice")

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	// First drain: dispatches the build, which blocks in the gate.
	go db.indexBuildWorker.drain(ctx)
	<-entered

	// Second drain while the first build is in flight: the in-flight guard must skip it, so no
	// second build starts. drain returns immediately (dispatch only).
	db.indexBuildWorker.drain(ctx)

	// Give any erroneously-dispatched second build a chance to enter the gate.
	select {
	case <-entered:
		t.Fatal("a second build started for the same index despite the in-flight guard")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	db.indexBuildWorker.builds.Wait()

	assert.Equal(t, int32(1), starts.Load(), "exactly one build must run for the index")
	requireNoIndexState(t, ctx, db, col.Version().CollectionID, desc.ID)
}

// TestIndexWorker_Notify_Coalesces checks that notify never blocks and that multiple notifies
// while a drain is pending collapse to a single buffered wake.
func TestIndexWorker_Notify_Coalesces(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)
	w := db.indexBuildWorker

	// Several notifies in a row must not block; the size-1 buffer coalesces them.
	for range 5 {
		w.notify()
	}

	// Exactly one wake is buffered.
	select {
	case <-w.wake:
	default:
		t.Fatal("expected one buffered wake after notify")
	}
	select {
	case <-w.wake:
		t.Fatal("notify did not coalesce: more than one wake buffered")
	default:
	}
}

// TestIndexWorker_EventDrivenBuild_Completes checks the end-to-end async path with the background
// run loop active: creating an index wakes the worker via the published build event, and the index
// becomes ready without an explicit drain.
func TestIndexWorker_EventDrivenBuild_Completes(t *testing.T) {
	ctx := context.Background()
	// newBadgerDB starts the background worker.
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice"}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))

	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
	})
	require.NoError(t, err)

	// The build runs in the background; poll the persisted state until it clears (ready).
	collectionID := col.Version().CollectionID
	require.Eventually(t, func() bool {
		rawTxn, err := db.NewTxn(true)
		require.NoError(t, err)
		defer rawTxn.Discard()
		_, err = getIndexState(InitContext(ctx, rawTxn), collectionID, desc.ID)
		// A missing record means ready.
		return err != nil
	}, 10*time.Second, 5*time.Millisecond, "index did not become ready via the event-driven worker")

	require.Len(t, queryUserByName(t, db, ctx, "Alice"), 1)
}

// TestIndexWorker_DrainDroppingRecord_Completes checks that a drain over a committed dropping
// record GCs the index entries and clears the record.
func TestIndexWorker_DrainDroppingRecord_Completes(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	for _, name := range []string{"Alice", "Bob", "Carol"} {
		addUserDoc(t, ctx, col, name)
	}

	desc, err := newNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	collectionID := col.Version().CollectionID
	shortID := getCollectionShortID(t, ctx, db, collectionID)
	require.Equal(t, 3, countIndexEntries(t, ctx, db, shortID, desc.ID))

	// Stage a dropping record on its own transaction, then drain.
	require.NoError(t, db.withTxnRetries(ctx, func(c context.Context) error {
		return db.startIndexDrop(c, collectionID, desc.ID)
	}))

	db.indexBuildWorker.drainSync(ctx)

	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))
	requireNoIndexState(t, ctx, db, collectionID, desc.ID)
}

// TestIndexWorker_ConcurrentBuilds_AllComplete checks that several indexes building at once all
// complete under the bounded pool.
func TestIndexWorker_ConcurrentBuilds_AllComplete(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, `type User { name: String, age: Int, score: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":30,"score":9}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc))

	var wg sync.WaitGroup
	descs := make([]client.IndexDescription, 3)
	fields := []string{"name", "age", "score"}
	for i, f := range fields {
		desc, err := col.NewIndex(ctx, client.NewIndexRequest{
			Fields: []client.IndexedFieldDescription{{Name: f}},
		})
		require.NoError(t, err)
		descs[i] = desc
	}
	wg.Wait()

	collectionID := col.Version().CollectionID
	for _, desc := range descs {
		desc := desc
		require.Eventually(t, func() bool {
			rawTxn, err := db.NewTxn(true)
			require.NoError(t, err)
			defer rawTxn.Discard()
			_, err = getIndexState(InitContext(ctx, rawTxn), collectionID, desc.ID)
			return err != nil
		}, 10*time.Second, 5*time.Millisecond, "index %d did not become ready", desc.ID)
	}
}

// TestInFlightKey_BuildAndDropDistinct checks that the in-flight key separates a build and a drop
// of the same index, so they are never treated as the same in-flight unit.
func TestInFlightKey_BuildAndDropDistinct(t *testing.T) {
	key := keys.IndexStateKey{CollectionID: "col1", IndexID: 7}
	build := inFlightKey(key, client.BackfillIndexAction)
	drop := inFlightKey(key, client.DropIndexAction)
	assert.NotEqual(t, build, drop)
}
