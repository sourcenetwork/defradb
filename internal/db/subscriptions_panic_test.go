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

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/event"
)

// A panic anywhere inside the selection-evaluation path of a subscription
// must not bring down the whole process — otherwise one bad subscription
// takes every other client's subscriptions and in-flight requests with
// it. The handler's goroutine has a recover() guarding that. This test
// swaps in a selection function that panics on every event, drives an
// event through, and confirms the DB is still usable afterwards.
func TestSubscription_PanicInSelection_DoesNotKillDB(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	// Replace the selection step with one that always panics. Restore
	// the original on test exit so we don't poison sibling tests.
	original := runSubscriptionSelection
	defer func() { runSubscriptionSelection = original }()
	runSubscriptionSelection = func(
		_ context.Context,
		_ *DB,
		_ request.Selection,
	) (map[string]any, error) {
		panic("synthetic panic from test")
	}

	res := db.ExecRequest(ctx, `subscription {
		User {
			_docID
			name
		}
	}`)
	require.Empty(t, res.GQL.Errors)
	require.NotNil(t, res.Subscription)

	// Fire an event the subscription will pass through its docID/cid
	// filters and hand to the (now-panicking) selection step. A docID
	// of "" matches the subscription's default filter, and any valid
	// CID works since we never reach the planner.
	bogusCid, err := cid.Decode("bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m")
	require.NoError(t, err)

	db.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID: "bae-00000000-0000-0000-0000-000000000000",
		Cid:   bogusCid,
	}))

	// Give the goroutine a beat to handle the event. If the recover
	// isn't there, the test binary dies here with exit code 2 and the
	// `require` below never runs.
	time.Sleep(50 * time.Millisecond)

	// The DB must still be alive. Run an unrelated query through the
	// same instance and confirm it returns cleanly.
	check := db.ExecRequest(ctx, `query { User { _docID } }`)
	require.Empty(t, check.GQL.Errors, "DB must still be usable after a subscription panic")
}
