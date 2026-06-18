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
	require.NoError(t, db.setIndexState(ctx, "col1", 1, indexState{Status: client.IndexStatusBuilding}))

	// ... alongside a collection-wide action record under the same collection prefix.
	require.NoError(t, action.SetTxn(
		ctx, db.events, "col1", client.TruncateAction, "",
		client.InProgressActionStatus, "", nil, false,
	))

	states, err := getIndexStates(ctx, "col1")
	require.NoError(t, err)
	require.Len(t, states, 1, "only the per-subject index record must be returned")
	assert.Equal(t, client.IndexStatusBuilding, states[1].Status)
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

	require.NoError(t, db.setIndexState(ctx, "col1", 1, indexState{Status: client.IndexStatusBuilding}))

	// Write a record under the same collection prefix whose value is neither valid JSON nor a
	// valid uvarint, i.e. a corrupt index action record.
	txn := datastore.CtxMustGetTxn(ctx)
	corruptKey := keys.NewActionStatusSubjectKey("col1", client.BackfillIndexAction, "2").Bytes()
	require.NoError(t, txn.Systemstore().Set(ctx, corruptKey, []byte{}))

	// Lenient path: the healthy record survives, the corrupt one is skipped.
	states, err := getIndexStates(ctx, "col1")
	require.NoError(t, err)
	require.Len(t, states, 1)
	assert.Equal(t, client.IndexStatusBuilding, states[1].Status)

	// Strict path: recovery surfaces the corruption.
	_, err = listIndexStates(ctx)
	require.Error(t, err)
}

// TestToIndexState_NonIndexStatesAreSkipped checks that action/status combinations that do not
// describe a live index lifecycle state report ok=false (and so are skipped by scans).
func TestToIndexState_NonIndexStatesAreSkipped(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action client.Action
		status client.ActionStatus
	}{
		{"backfill completed", client.BackfillIndexAction, client.CompletedActionStatus},
		{"backfill none", client.BackfillIndexAction, client.NoneActionStatus},
		{"drop errored", client.DropIndexAction, client.ErroredActionStatus},
		{"drop completed", client.DropIndexAction, client.CompletedActionStatus},
		{"truncate in progress", client.TruncateAction, client.InProgressActionStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec := client.ActionExecution{Action: tc.action, Status: tc.status}
			_, ok := toIndexState(exec, indexBackfillPayload{}, "")
			assert.False(t, ok, "%s must not project to an index state", tc.name)
		})
	}
}
