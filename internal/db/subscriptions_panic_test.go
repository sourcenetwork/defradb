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

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/event"
)

// newPanickingProcessEventFixture sets up the inputs processEvent needs
// (DB + collection + bogus event + minimal subRequest) so tests can
// drive it directly with an injected selectFn.
func newPanickingProcessEventFixture(t *testing.T) (*DB, subscriptionSelector, event.Update, chan client.GQLResult) {
	t.Helper()
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	subReq := &request.Select{
		Field: request.Field{Name: "User"},
	}

	bogusCid, err := cid.Decode("bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m")
	require.NoError(t, err)

	evt := event.Update{
		DocID: "bae-00000000-0000-0000-0000-000000000000",
		Cid:   bogusCid,
	}

	resCh := make(chan client.GQLResult, 4)
	return db, subReq, evt, resCh
}

// A panic in a subscription's selection eval must be recovered, so the
// caller of processEvent doesn't die.
func TestProcessEvent_PanicInSelection_IsRecovered(t *testing.T) {
	db, subReq, evt, resCh := newPanickingProcessEventFixture(t)

	selectFn := func(_ context.Context, _ *DB, _ request.Selection) (map[string]any, error) {
		panic("synthetic panic from test")
	}

	require.NotPanics(t, func() {
		processEvent(context.Background(), db, subReq, evt, resCh, selectFn)
	})
}

// A panic on one event must not affect a subsequent event's delivery:
// processEvent for the second event still pushes its result to resCh.
func TestProcessEvent_PanicOnOne_DoesNotAffectNext(t *testing.T) {
	db, subReq, evt, resCh := newPanickingProcessEventFixture(t)

	panicking := func(_ context.Context, _ *DB, _ request.Selection) (map[string]any, error) {
		panic("synthetic panic on first event")
	}
	working := func(_ context.Context, _ *DB, _ request.Selection) (map[string]any, error) {
		return map[string]any{
			"User": []map[string]any{
				{"_docID": "bae-test", "name": "Alice"},
			},
		}, nil
	}

	require.NotPanics(t, func() {
		processEvent(context.Background(), db, subReq, evt, resCh, panicking)
	})

	processEvent(context.Background(), db, subReq, evt, resCh, working)

	select {
	case result, ok := <-resCh:
		require.True(t, ok, "result channel must be open")
		require.NotNil(t, result.Data, "second event must deliver a result")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second event never delivered to resCh")
	}
}
