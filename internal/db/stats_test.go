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
	"os"
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
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// hostSchema is the schema the Shinzo host deploys, copied verbatim. The counters here
// only mean anything against the collections and indexes production actually runs, and a
// hand-simplified schema has already produced two contradictory conclusions.
func hostSchema(t *testing.T) string {
	t.Helper()
	sdl, err := os.ReadFile("testdata/shinzo_host_schema.graphql")
	require.NoError(t, err)
	return string(sdl)
}

// The index layout is what makes the unique-index and read-kind counters meaningful, so
// this asserts the fixture is the layout production runs rather than trusting the file.
func TestHostSchemaIndexes(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, hostSchema(t))
	require.NoError(t, err)

	want := map[string]struct {
		indexes int
		unique  int
	}{
		"Ethereum__Mainnet__Block":             {indexes: 2, unique: 1},
		"Ethereum__Mainnet__Transaction":       {indexes: 3, unique: 1},
		"Ethereum__Mainnet__Log":               {indexes: 4, unique: 0},
		"Ethereum__Mainnet__AccessListEntry":   {indexes: 2, unique: 0},
		"Ethereum__Mainnet__BlockSignature":    {indexes: 1, unique: 0},
		"Ethereum__Mainnet__AttestationRecord": {indexes: 2, unique: 0},
		"Ethereum__Mainnet__SnapshotSignature": {indexes: 0, unique: 0},
	}

	total, totalUnique := 0, 0
	for name, expected := range want {
		col, err := db.GetCollectionByName(ctx, name)
		require.NoError(t, err, name)

		descs, err := col.ListIndexes(ctx)
		require.NoError(t, err, name)

		unique := 0
		for _, d := range descs {
			if d.Description.Unique {
				unique++
			}
		}
		require.Equal(t, expected.indexes, len(descs), "index count for %s", name)
		require.Equal(t, expected.unique, unique, "unique index count for %s", name)

		total += len(descs)
		totalUnique += unique
	}
	require.Equal(t, 14, total)
	require.Equal(t, 2, totalUnique)
}

// A merge of a document with no local heads is a create; a merge onto an existing head is
// an update. The split decides whether the range-iterator over-read can apply at all, so
// it is asserted rather than assumed.
func TestMergeStatsCreateVsUpdate(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, hostSchema(t))
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "Ethereum__Mainnet__BlockSignature")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	concrete, ok := col.(*collection)
	require.True(t, ok)

	first := map[string]any{"blockHash": "0xaa"}
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

	update, err := builder.generateCompositeUpdate(&lsys, map[string]any{"blockHash": "0xbb"}, create)
	require.NoError(t, err)

	require.NoError(t, db.executeMerge(ctx, concrete, event.Merge{
		DocID:        docID.String(),
		Cid:          update.link.Cid,
		CollectionID: col.Version().CollectionID,
	}))
	require.Equal(t, int64(1), db.stats.creates.Load())
	require.Equal(t, int64(1), db.stats.updates.Load())
}

// Block.hash is unique, so two blocks carrying the same hash collide on the existence
// check saveUniqueKey performs. The counter separates a collision from a plain conflict:
// a collision fails the write outright rather than exhausting the retries.
func TestMergeStatsUniqueIndexHit(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, hostSchema(t))
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "Ethereum__Mainnet__Block")
	require.NoError(t, err)

	first, err := client.NewDocFromMap(ctx, map[string]any{"hash": "0xdup", "number": 1}, col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, first))
	require.Equal(t, int64(1), db.stats.uniqueIndexChecks.Load())
	require.Equal(t, int64(0), db.stats.uniqueIndexHits.Load())

	second, err := client.NewDocFromMap(ctx, map[string]any{"hash": "0xdup", "number": 2}, col.Version())
	require.NoError(t, err)
	require.Error(t, col.AddDocument(ctx, second))
	require.Equal(t, int64(2), db.stats.uniqueIndexChecks.Load())
	require.Equal(t, int64(1), db.stats.uniqueIndexHits.Load())
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
			name: "unique index violation",
			err:  NewErrMergeEventDropped(errors.New(errCanNotIndexNonUniqueFields), "bae-1", "bafy"),
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
