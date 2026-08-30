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
	"testing"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/errors"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

func TestMergeStatsCreateVsUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	concrete, ok := col.(*collection)
	require.True(t, ok)

	first := map[string]any{"name": "John"}
	builder, _ := newDagBuilder(ctx, col, first)

	create, err := builder.generateCompositeUpdate(&lsys, first, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(create.link.Cid)

	require.NoError(t, db.executeMerge(ctx, concrete, event.Merge{
		DocID:        docID.String(),
		Cid:          create.link.Cid,
		CollectionID: col.Version().CollectionID,
	}))
	require.Equal(t, int64(1), db.stats.creates.Load())
	require.Equal(t, int64(0), db.stats.updates.Load())

	update, err := builder.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, create)
	require.NoError(t, err)

	require.NoError(t, db.executeMerge(ctx, concrete, event.Merge{
		DocID:        docID.String(),
		Cid:          update.link.Cid,
		CollectionID: col.Version().CollectionID,
	}))
	require.Equal(t, int64(1), db.stats.creates.Load())
	require.Equal(t, int64(1), db.stats.updates.Load())
}

// A dropped event is a document this node did not store. The count alone does not say what
// to do about it: a missing block is the sender's DAG being short, a unique index violation
// is two documents claiming one value, and retry exhaustion is contention.
func TestMergeDropReason(t *testing.T) {
	missing := ipld.ErrNotFound{Cid: cid.Undef}

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{
			name: "missing child block, wrapped as the merge path wraps it",
			err:  NewErrMergeEventDropped(NewErrLoadChildBlock(missing, "bafy"), "bae-1", "bafy"),
			want: dropMissingBlock,
		},
		{
			name: "unique index violation, carrying the fields the index path attaches",
			err: NewErrMergeEventDropped(
				NewErrCanNotIndexNonUniqueFields("bae-1", errors.NewKV("name", "John")),
				"bae-1",
				"bafy",
			),
			want: dropUniqueIndex,
		},
		{
			name: "retry exhaustion",
			err:  NewErrMergeEventDropped(client.NewErrMaxTxnRetries(corekv.ErrTxnConflict), "bae-1", "bafy"),
			want: dropRetryExhausted,
		},
		{
			name: "anything else is not silently filed under a known cause",
			err:  errors.New("something new"),
			want: dropOther,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, mergeDropReason(tc.err))
		})
	}
}

// The reasons have to survive the drain that reports them, and reset afterwards so each
// line carries the interval rather than a running total.
func TestMergeStatsDrainDropReasons(t *testing.T) {
	var s mergeStats
	s.markDropped(dropMissingBlock)
	s.markDropped(dropMissingBlock)
	s.markDropped(dropUniqueIndex)

	attrs := s.drainDropReasons()
	got := map[string]int64{}
	for _, a := range attrs {
		got[a.Key] = a.Value.Int64()
	}
	require.Equal(t, map[string]int64{dropMissingBlock: 2, dropUniqueIndex: 1}, got)

	require.Nil(t, s.drainDropReasons(), "a drained interval must not repeat its counts")
}

// A chunk that fails is re-run one event at a time, so an event that merged in the failed
// attempt is merged again on the isolation pass. Only the merge that commits is counted.
func TestMergeStatsCountsCommittedMergeOnce(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	goodState := map[string]any{"name": "John"}
	goodBuilder, _ := newDagBuilder(ctx, col, goodState)
	good, err := goodBuilder.generateCompositeUpdate(&lsys, goodState, compositeInfo{})
	require.NoError(t, err)
	goodDocID := client.NewDocIDV0(good.link.Cid)

	// A second document whose parent composite was never stored, so it cannot merge.
	badState := map[string]any{"name": "Jane"}
	badBuilder, _ := newDagBuilder(ctx, col, badState)
	badGenesis, err := badBuilder.generateCompositeUpdate(&lsys, badState, compositeInfo{})
	require.NoError(t, err)
	badDocID := client.NewDocIDV0(badGenesis.link.Cid)

	missingBlock := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Status: 1}}}
	missingLink, err := coreblock.GetLinkFromNode(missingBlock.GenerateNode())
	require.NoError(t, err)

	bad, err := badBuilder.generateCompositeUpdate(
		&lsys,
		map[string]any{"name": "Janet"},
		compositeInfo{link: missingLink, height: 2},
	)
	require.NoError(t, err)

	// The mergeable event is listed first so it lands in the chunk attempt that then fails,
	// and again when the chunk is re-run one event at a time.
	merged, err := db.MergeBatchWithTxn(ctx, []event.Merge{
		{DocID: goodDocID.String(), Cid: good.link.Cid, CollectionID: col.CollectionID()},
		{DocID: badDocID.String(), Cid: bad.link.Cid, CollectionID: col.CollectionID()},
	})
	require.Error(t, err)
	require.Equal(t, []bool{true, false}, merged)

	require.Equal(t, int64(1), db.stats.creates.Load(), "the committed merge is counted once")
	require.Equal(t, int64(0), db.stats.updates.Load())
}

// existingDocUpdates creates count documents and returns an update event for each. They are
// committed up front so the caller's budget can only fail the merge itself: creating a document
// commits the short ID it reserves first, which would spend the budget before the merge.
func existingDocUpdates(ctx context.Context, t *testing.T, db *DB, col client.Collection, count int) []event.Merge {
	t.Helper()

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	updates := make([]event.Merge, count)
	for i := range updates {
		state := map[string]any{"name": fmt.Sprintf("user-%d", i)}
		builder, _ := newDagBuilder(ctx, col, state)

		genesis, err := builder.generateCompositeUpdate(&lsys, state, compositeInfo{})
		require.NoError(t, err)
		docID := client.NewDocIDV0(genesis.link.Cid)
		require.NoError(t, db.Merge(ctx, event.Merge{
			DocID:        docID.String(),
			Cid:          genesis.link.Cid,
			CollectionID: col.CollectionID(),
		}))

		update, err := builder.generateCompositeUpdate(&lsys, map[string]any{"name": "renamed"}, genesis)
		require.NoError(t, err)
		updates[i] = event.Merge{
			DocID:        docID.String(),
			Cid:          update.link.Cid,
			CollectionID: col.CollectionID(),
		}
	}
	return updates
}

// drainedDrops drains the drop reasons into a map.
func drainedDrops(s *mergeStats) map[string]int64 {
	drops := map[string]int64{}
	for _, attr := range s.drainDropReasons() {
		drops[attr.Key] = attr.Value.Int64()
	}
	return drops
}

// The single-event path has no fallback, so exhausting its retry budget is a dropped
// document rather than chunk exhaustion.
func TestMergeStatsSingleDocumentExhaustion(t *testing.T) {
	ctx := context.Background()
	db, store, err := newConflictingBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	updates := existingDocUpdates(ctx, t, db, col, 1)

	store.failCommits.Store(int64(db.txnAttempts()))
	err = db.Merge(ctx, updates[0])
	require.ErrorIs(t, err, client.ErrMaxTxnRetries)
	require.ErrorContains(t, err, "DocID: "+updates[0].DocID, "the error must name the document")
	require.ErrorContains(t, err, "CID: "+updates[0].Cid.String(), "the error must name the commit")

	require.Equal(t, int64(db.txnAttempts()), db.stats.txnConflicts.Load(),
		"each abandoned transaction is one conflict")
	require.Equal(t, int64(0), db.stats.chunkExhausted.Load(),
		"there is no chunk on this path")
	require.Equal(t, map[string]int64{dropRetryExhausted: 1}, drainedDrops(db.stats))
}

// A chunk that runs out of attempts is counted once, whether or not the per-event re-runs
// that follow succeed. A chunk of one event has no smaller write set and is not counted.
func TestMergeStatsChunkExhaustion(t *testing.T) {
	attempts := int64(defaultMaxTxnRetries)

	for _, tc := range []struct {
		name string
		// failCommits is spent in full by every case, so it also states how many transactions
		// the batch and isolation passes together are expected to abandon.
		failCommits   int64
		documents     int
		wantMerged    bool
		wantExhausted int64
		wantDrops     map[string]int64
	}{
		{
			name:          "a chunk is counted when it exhausts, even though its re-runs recover",
			failCommits:   attempts,
			documents:     mergeChunkSize,
			wantMerged:    true,
			wantExhausted: 1,
			wantDrops:     map[string]int64{},
		},
		{
			name:          "a chunk is counted once however many of its re-runs are lost",
			failCommits:   attempts * (1 + mergeChunkSize),
			documents:     mergeChunkSize,
			wantMerged:    false,
			wantExhausted: 1,
			wantDrops:     map[string]int64{dropRetryExhausted: mergeChunkSize},
		},
		{
			name:          "a chunk of one is not counted",
			failCommits:   attempts * 2,
			documents:     1,
			wantMerged:    false,
			wantExhausted: 0,
			wantDrops:     map[string]int64{dropRetryExhausted: 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, store, err := newConflictingBadgerDB(ctx)
			require.NoError(t, err)
			t.Cleanup(db.Close)

			_, err = db.AddCollection(ctx, userSchema)
			require.NoError(t, err)
			col, err := db.GetCollectionByName(ctx, "User")
			require.NoError(t, err)

			updates := existingDocUpdates(ctx, t, db, col, tc.documents)

			store.failCommits.Store(tc.failCommits)
			merged, err := db.MergeBatchWithTxn(ctx, updates)
			if tc.wantMerged {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Len(t, merged, tc.documents)
			for i, ok := range merged {
				require.Equal(t, tc.wantMerged, ok, "event %d", i)
			}

			require.Equal(t, tc.failCommits, db.stats.txnConflicts.Load(),
				"every abandoned transaction is one conflict")
			require.Equal(t, tc.wantExhausted, db.stats.chunkExhausted.Load())
			require.Equal(t, tc.wantDrops, drainedDrops(db.stats))
		})
	}
}

// The collection lookup can fail without the collection being missing.
func TestMergeStatsCollectionLookupCauses(t *testing.T) {
	newMergeEvent := func(ctx context.Context, t *testing.T, db *DB, collectionID string) event.Merge {
		t.Helper()

		col, err := db.GetCollectionByName(ctx, "User")
		require.NoError(t, err)

		lsys := cidlink.DefaultLinkSystem()
		lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

		state := map[string]any{"name": "John"}
		builder, _ := newDagBuilder(ctx, col, state)
		genesis, err := builder.generateCompositeUpdate(&lsys, state, compositeInfo{})
		require.NoError(t, err)

		return event.Merge{
			DocID:        client.NewDocIDV0(genesis.link.Cid).String(),
			Cid:          genesis.link.Cid,
			CollectionID: collectionID,
		}
	}

	t.Run("a collection this node does not hold", func(t *testing.T) {
		ctx := context.Background()
		db, err := newBadgerDB(ctx)
		require.NoError(t, err)
		t.Cleanup(db.Close)

		_, err = db.AddCollection(ctx, userSchema)
		require.NoError(t, err)

		require.ErrorIs(t, db.Merge(ctx, newMergeEvent(ctx, t, db, "not-a-collection")),
			client.ErrCollectionNotFound)
		require.Equal(t, map[string]int64{dropCollection: 1}, drainedDrops(db.stats))
	})

	// closedDB returns a database that can no longer open a transaction, and an event for a
	// collection it holds, so the lookup fails before reading any collection.
	closedDB := func(t *testing.T) (*DB, event.Merge) {
		t.Helper()

		ctx := context.Background()
		db, err := newBadgerDB(ctx)
		require.NoError(t, err)
		t.Cleanup(db.Close)

		_, err = db.AddCollection(ctx, userSchema)
		require.NoError(t, err)
		col, err := db.GetCollectionByName(ctx, "User")
		require.NoError(t, err)

		evt := newMergeEvent(ctx, t, db, col.CollectionID())
		db.ctxCancel()
		return db, evt
	}

	t.Run("a lookup that fails for any other reason", func(t *testing.T) {
		db, evt := closedDB(t)

		require.ErrorIs(t, db.Merge(context.Background(), evt), context.Canceled)
		require.Equal(t, map[string]int64{dropContext: 1}, drainedDrops(db.stats))
	})

	t.Run("a lookup that fails for any other reason, in a batch", func(t *testing.T) {
		db, evt := closedDB(t)

		merged, err := db.MergeBatchWithTxn(context.Background(), []event.Merge{evt})
		require.ErrorIs(t, err, context.Canceled)
		require.Equal(t, []bool{false}, merged)
		require.Equal(t, map[string]int64{dropContext: 1}, drainedDrops(db.stats))
	})
}
