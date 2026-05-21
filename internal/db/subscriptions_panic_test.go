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

// A panic in a subscription's selection eval must not kill the process.
func TestSubscription_PanicInSelection_DoesNotKillDB(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

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

	bogusCid, err := cid.Decode("bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m")
	require.NoError(t, err)

	db.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID: "bae-00000000-0000-0000-0000-000000000000",
		Cid:   bogusCid,
	}))

	time.Sleep(50 * time.Millisecond)

	check := db.ExecRequest(ctx, `query { User { _docID } }`)
	require.Empty(t, check.GQL.Errors, "DB must still be usable after a subscription panic")
}

// A panic on one event must not close the subscription stream:
// subsequent events still need to deliver.
func TestSubscription_PanicInOneEvent_DoesNotEndStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	original := runSubscriptionSelection
	defer func() { runSubscriptionSelection = original }()

	var callCount int
	runSubscriptionSelection = func(
		_ context.Context,
		_ *DB,
		_ request.Selection,
	) (map[string]any, error) {
		callCount++
		if callCount == 1 {
			panic("synthetic panic on first event")
		}
		return map[string]any{
			"User": []map[string]any{
				{"_docID": "bae-test", "name": "Alice"},
			},
		}, nil
	}

	res := db.ExecRequest(ctx, `subscription {
		User {
			_docID
			name
		}
	}`)
	require.Empty(t, res.GQL.Errors)
	require.NotNil(t, res.Subscription)

	bogusCid, err := cid.Decode("bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m")
	require.NoError(t, err)

	// First event panics in selection; recover should catch it.
	db.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID: "bae-00000000-0000-0000-0000-000000000000",
		Cid:   bogusCid,
	}))
	time.Sleep(50 * time.Millisecond)

	// Second event returns a real result; must still be delivered.
	db.events.Publish(event.NewMessage(event.UpdateName, event.Update{
		DocID: "bae-11111111-1111-1111-1111-111111111111",
		Cid:   bogusCid,
	}))

	select {
	case result, ok := <-res.Subscription:
		require.True(t, ok, "subscription channel must still be open after a recovered panic")
		require.NotNil(t, result.Data, "second event must deliver a result on the same stream")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("subscription stream was closed by the recovered panic; second event never delivered")
	}
}
