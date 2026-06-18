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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/action"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// drainActionEvents collects ActionExecution events from a subscription until no event
// arrives within a short quiet period, then returns them.
func drainActionEvents(t *testing.T, sub event.Subscription) []event.ActionExecution {
	t.Helper()
	var got []event.ActionExecution
	for {
		select {
		case msg, ok := <-sub.Message():
			if !ok {
				return got
			}
			exec, ok := msg.Data.(event.ActionExecution)
			require.True(t, ok, "unexpected event payload %T", msg.Data)
			got = append(got, exec)
		case <-time.After(200 * time.Millisecond):
			return got
		}
	}
}

// TestBackfillIndex_Completion_PublishesNoDropEvent guards against the regression where
// completing a build also completed the (non-existent) drop record, emitting a spurious
// DropIndexAction/Completed event. A successful build must only ever emit BackfillIndexAction
// events.
func TestBackfillIndex_Completion_PublishesNoDropEvent(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	addUserDoc(t, ctx, col, "Alice")

	sub, err := db.events.Subscribe(event.ActionExecutionName)
	require.NoError(t, err)
	defer db.events.Unsubscribe(sub)

	_, err = newNameIndex(t, ctx, col)
	require.NoError(t, err)

	for _, exec := range drainActionEvents(t, sub) {
		assert.NotEqual(t, client.DropIndexAction, exec.Action,
			"a backfill completion must not publish a DropIndexAction event")
	}
}

// TestBackfillIndex_MultiBatch_PublishesSingleBuildingEvent checks the event-suppression
// invariant: a multi-batch build emits exactly one building event (the initial transition),
// not one per batch, so per-batch watermark advances do not flood the bus.
func TestBackfillIndex_MultiBatch_PublishesSingleBuildingEvent(t *testing.T) {
	origBatchSize := indexBackfillBatchSize
	indexBackfillBatchSize = 2
	defer func() { indexBackfillBatchSize = origBatchSize }()

	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	for i := range 7 {
		addUserDoc(t, ctx, col, "name"+string(rune('0'+i)))
	}

	sub, err := db.events.Subscribe(event.ActionExecutionName)
	require.NoError(t, err)
	defer db.events.Unsubscribe(sub)

	_, err = newNameIndex(t, ctx, col)
	require.NoError(t, err)

	building := 0
	for _, exec := range drainActionEvents(t, sub) {
		if exec.Action == client.BackfillIndexAction && exec.Status == client.InProgressActionStatus {
			building++
		}
	}
	assert.Equal(t, 1, building,
		"a multi-batch build must publish exactly one building event regardless of batch count")
}

// TestScanIndexStates_IgnoresCollectionWideActions checks that a collection-wide action record
// (empty subject, e.g. truncate) sharing the collection prefix is excluded from the index state
// map, so it cannot be misread as an index lifecycle state.
func TestScanIndexStates_IgnoresCollectionWideActions(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	// An index state record (per-subject) ...
	require.NoError(t, db.startIndexBuild(ctx, "col1", 1))

	// ... alongside a collection-wide action record under the same collection prefix.
	require.NoError(t, action.SetTxn(
		ctx, db.events, "col1", client.TruncateAction, "",
		client.InProgressActionStatus, "", nil, false,
	))

	states, err := getIndexStates(ctx, "col1")
	require.NoError(t, err)
	require.Len(t, states, 1, "only the per-subject index record must be returned")
	assert.True(t, states[1].isBuilding())
}

// TestListActions_AfterFailedBuild_ReportsErroredRecordWithSubject checks the observability
// guarantee for a permanently failed index: the errored action record survives, and ListActions
// reports it with the index ID as Subject and an errored status. (A successful build deletes its
// record; only failed/in-flight records remain.)
func TestListActions_AfterFailedBuild_ReportsErroredRecordWithSubject(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, "type User { name: String\n age: Int }")
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// Two docs sharing an age make the unique backfill fail (non-retryable).
	doc1, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice","age":21}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc1))
	doc2, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Bob","age":21}`), col.Version())
	require.NoError(t, err)
	require.NoError(t, col.AddDocument(ctx, doc2))

	// The implicit-txn path returns a zero desc on backfill failure, so the index ID is read
	// back from the listing (the definition persists with a failed status).
	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
		Unique: true,
	})
	require.Error(t, err)

	indexes, err := col.ListIndexes(ctx)
	require.NoError(t, err)
	require.Len(t, indexes, 1)
	failedIndexID := indexes[0].ID

	actions, err := db.ListActions(ctx)
	require.NoError(t, err)

	require.Len(t, actions, 1, "the failed build must leave exactly one errored action record")
	assert.Equal(t, client.BackfillIndexAction, actions[0].Action)
	assert.Equal(t, client.ErroredActionStatus, actions[0].Status)
	assert.Equal(t, indexSubject(failedIndexID), actions[0].Subject,
		"the errored record must be keyed by the failed index ID")
	assert.Equal(t, col.Version().CollectionID, actions[0].CollectionID)
}

// TestGetIndexStates_SkipsCorruptRecord checks that a single undecodable action record does not
// fail the collection-open path: the healthy record is still returned and the corrupt one is
// skipped. listIndexStates (recovery) stays strict and surfaces the error.
func TestGetIndexStates_SkipsCorruptRecord(t *testing.T) {
	db, ctx, cleanup := newIndexStateTestCtx(t)
	defer cleanup()

	require.NoError(t, db.startIndexBuild(ctx, "col1", 1))

	// Index 2 has a valid status record but a payload that is not valid JSON, i.e. corrupt.
	require.NoError(t, db.startIndexBuild(ctx, "col1", 2))
	txn := datastore.CtxMustGetTxn(ctx)
	payloadKey := keys.NewActionPayloadKey("col1", client.BackfillIndexAction, "2").Bytes()
	require.NoError(t, txn.Systemstore().Set(ctx, payloadKey, []byte("not json")))

	// Lenient path: the healthy record survives, the corrupt one is skipped.
	states, err := getIndexStates(ctx, "col1")
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.True(t, states[1].isBuilding())

	// Strict path: recovery surfaces the corruption.
	_, err = listIndexStates(ctx)
	require.Error(t, err)
}

// TestIndexState_Predicates classifies (Action, Status) pairs into the building/failed/dropping
// lifecycle predicates, and confirms unrelated combinations match none of them.
func TestIndexState_Predicates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		state    indexState
		building bool
		failed   bool
		dropping bool
	}{
		{"building", indexState{Action: client.BackfillIndexAction, Status: client.InProgressActionStatus}, true, false, false},
		{"failed", indexState{Action: client.BackfillIndexAction, Status: client.ErroredActionStatus}, false, true, false},
		{"dropping", indexState{Action: client.DropIndexAction, Status: client.InProgressActionStatus}, false, false, true},
		{"backfill completed", indexState{Action: client.BackfillIndexAction, Status: client.CompletedActionStatus}, false, false, false},
		{"drop errored", indexState{Action: client.DropIndexAction, Status: client.ErroredActionStatus}, false, false, false},
		{"truncate", indexState{Action: client.TruncateAction, Status: client.InProgressActionStatus}, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.building, tc.state.isBuilding())
			assert.Equal(t, tc.failed, tc.state.isFailed())
			assert.Equal(t, tc.dropping, tc.state.isDropping())
		})
	}
}
