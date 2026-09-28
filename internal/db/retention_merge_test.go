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

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	intOpts "github.com/sourcenetwork/defradb/internal/options"
)

const retentionLogSchema = `type Log { blockNumber: Int  address: String }`

// floorRule covers every collection with one height field and floor, or none when field is empty.
type floorRule struct {
	field string
	floor int64
}

func (r floorRule) RetentionFloor(string) (string, int64, bool) {
	return r.field, r.floor, r.field != ""
}

func newRetentionDB(t *testing.T, rule client.RetentionRule, schema string) (*DB, client.Collection) {
	t.Helper()
	ctx := context.Background()
	db, err := newBadgerDB(ctx, intOpts.DB().SetRetentionRule(rule))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	_, err = db.AddCollection(ctx, schema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "Log")
	require.NoError(t, err)
	return db, col
}

// stagingLinkSystem writes blocks with to-merge markers, as fetched blocks are stored.
func stagingLinkSystem(db *DB) *linking.LinkSystem {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.P2PBlockstoreFrom(db.rootstore, immutable.None[int]())))
	return &lsys
}

// stageLog stages a Log's blocks and returns the event that merges it.
func stageLog(t *testing.T, db *DB, col client.Collection, fields map[string]any) (client.DocID, event.Merge) {
	t.Helper()
	builder, _ := newDagBuilder(context.Background(), col, fields)
	head, err := builder.generateCompositeUpdate(stagingLinkSystem(db), fields, compositeInfo{})
	require.NoError(t, err)

	docID := client.NewDocIDV0(head.link.Cid)
	return docID, event.Merge{DocID: docID.String(), Cid: head.link.Cid, CollectionID: col.CollectionID()}
}

func isStored(t *testing.T, col client.Collection, docID client.DocID) bool {
	t.Helper()
	_, err := col.GetDocument(context.Background(), docID)
	if err != nil {
		require.ErrorIs(t, err, client.ErrDocumentNotFoundOrNotAuthorized)
		return false
	}
	return true
}

// The merge refuses what the rule refuses. A refusal is counted by reason, not as a drop, and the
// head is not marked merged.
func TestMergeAppliesTheRetentionRule(t *testing.T) {
	floor100 := floorRule{field: "blockNumber", floor: 100}
	noFloorYet := floorRule{field: "blockNumber"}

	for _, tc := range []struct {
		name    string
		rule    client.RetentionRule
		fields  map[string]any
		wantErr error
	}{
		{name: "at the floor", rule: floor100, fields: map[string]any{"blockNumber": 100}, wantErr: client.ErrBelowRetentionFloor},
		{name: "above the floor", rule: floor100, fields: map[string]any{"blockNumber": 101}},
		{name: "no height", rule: floor100, fields: map[string]any{"address": "0xa"}, wantErr: client.ErrNoRetentionHeight},
		{name: "negative height", rule: floor100, fields: map[string]any{"blockNumber": -5}, wantErr: client.ErrNoRetentionHeight},
		{name: "no floor yet", rule: noFloorYet, fields: map[string]any{"blockNumber": 0}},
		{name: "no floor yet and no height", rule: noFloorYet, fields: map[string]any{"address": "0xa"}, wantErr: client.ErrNoRetentionHeight},
		{name: "collection not covered", rule: floorRule{}, fields: map[string]any{"address": "0xa"}},
		{name: "no rule", fields: map[string]any{"address": "0xa"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, col := newRetentionDB(t, tc.rule, retentionLogSchema)
			docID, evt := stageLog(t, db, col, tc.fields)

			err := db.Merge(ctx, evt)

			if tc.wantErr == nil {
				require.NoError(t, err)
				require.True(t, isStored(t, col, docID))
				return
			}
			require.ErrorIs(t, err, tc.wantErr)
			require.False(t, isStored(t, col, docID))

			merged, err := datastore.P2PBlockstoreFrom(db.rootstore, immutable.None[int]()).IsMerged(ctx, evt.Cid)
			require.NoError(t, err)
			require.False(t, merged, "a refused document's head is not marked merged")

			var wantBelowFloor, wantNoHeight int64
			if errors.Is(tc.wantErr, client.ErrNoRetentionHeight) {
				wantNoHeight = 1
			} else {
				wantBelowFloor = 1
			}
			require.Equal(t, wantBelowFloor, db.stats.rejectedBelowFloor.Load())
			require.Equal(t, wantNoHeight, db.stats.rejectedNoHeight.Load())
			require.Empty(t, db.stats.drainDropReasons(), "a refusal is not a drop")
		})
	}
}

// A refused event in a batch does not stop the rest of its chunk.
func TestMergeBatchRejectsOnlyTheRefusedEvent(t *testing.T) {
	ctx := context.Background()
	db, col := newRetentionDB(t, floorRule{field: "blockNumber", floor: 100}, retentionLogSchema)
	newID, newEvt := stageLog(t, db, col, map[string]any{"blockNumber": 101, "address": "0xa"})
	oldID, oldEvt := stageLog(t, db, col, map[string]any{"blockNumber": 100, "address": "0xb"})
	laterID, laterEvt := stageLog(t, db, col, map[string]any{"blockNumber": 150, "address": "0xc"})

	outcomes, err := db.MergeBatchWithTxn(ctx, []event.Merge{newEvt, oldEvt, laterEvt})

	require.NoError(t, err)
	require.Equal(t, []event.MergeOutcome{event.MergeCommitted, event.MergeRejected, event.MergeCommitted}, outcomes)
	require.True(t, isStored(t, col, newID))
	require.False(t, isStored(t, col, oldID))
	require.True(t, isStored(t, col, laterID))
	require.Equal(t, int64(1), db.stats.rejectedBelowFloor.Load())
	require.Empty(t, db.stats.drainDropReasons())
}

// A delete is not judged.
func TestMergeAppliesADeleteBelowTheFloor(t *testing.T) {
	ctx := context.Background()
	db, col := newRetentionDB(t, floorRule{field: "blockNumber", floor: 100}, retentionLogSchema)
	fields := map[string]any{"blockNumber": 5}
	lsys := stagingLinkSystem(db)
	builder, _ := newDagBuilder(ctx, col, fields)
	created, err := builder.generateCompositeUpdate(lsys, fields, compositeInfo{})
	require.NoError(t, err)
	deleted, err := builder.generateCompositeDelete(lsys, created)
	require.NoError(t, err)
	docID := client.NewDocIDV0(created.link.Cid)

	err = db.Merge(ctx, event.Merge{DocID: docID.String(), Cid: deleted.link.Cid, CollectionID: col.CollectionID()})

	require.NoError(t, err)
	merged, err := datastore.P2PBlockstoreFrom(db.rootstore, immutable.None[int]()).IsMerged(ctx, deleted.link.Cid)
	require.NoError(t, err)
	require.True(t, merged)
}

// Only covered, non-branchable collections are judged, and nothing that cannot be resolved.
func TestRetentionFloorOf(t *testing.T) {
	rule := floorRule{field: "blockNumber", floor: 100}
	for _, tc := range []struct {
		name    string
		rule    client.RetentionRule
		schema  string
		unknown bool
		wantOK  bool
	}{
		{name: "covered", rule: rule, schema: retentionLogSchema, wantOK: true},
		{name: "branchable", rule: rule, schema: `type Log @branchable { blockNumber: Int }`},
		{name: "unknown collection", rule: rule, schema: retentionLogSchema, unknown: true},
		{name: "no rule", schema: retentionLogSchema},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, col := newRetentionDB(t, tc.rule, tc.schema)
			id := col.CollectionID()
			if tc.unknown {
				id = "bafyreicollectionthatdoesnotexist"
			}

			field, floor, ok := db.retentionFloorOf(context.Background(), id)

			require.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				require.Equal(t, "blockNumber", field)
				require.Equal(t, int64(100), floor)
			}
		})
	}
}
