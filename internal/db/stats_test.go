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
