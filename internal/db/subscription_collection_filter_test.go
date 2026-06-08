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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

const bookAuthorSchema = `
type Book {
	title: String
	author: Author
	publisher: Publisher
}
type Author {
	name: String
	wrote: Book @primary
}
type Publisher {
	label: String
	published: Book @primary
}
`

// A Book subscription must NOT open a new transaction in response to an
// Author event. db.previousTxnID is an atomic counter incremented inside
// every db.NewTxn call (db.go:226-232); reading it lets us observe whether
// the subscription event loop reached its db.NewTxn call (subscriptions.go:74)
// for the wrong-collection event.
//
// Pre-fix:  counter advances by >=1 — the subscription opens its own txn
//
//	before the planner fails inside VersionedFetcher.merge().
//
// Post-fix: counter is unchanged — CheckCollectionFilter rejected the event
//
//	at the docID/CID filter site (subscriptions.go:71) before any
//	transaction was opened.
func TestHandleSubscription_WrongCollectionEvent_OpensNoTxn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, bookAuthorSchema)
	require.NoError(t, err)

	authorCol, err := db.GetCollectionByName(ctx, "Author")
	require.NoError(t, err)

	res := db.ExecRequest(
		ctx,
		`subscription { Book { _docID title author { name } publisher { label } } }`,
	)
	require.Empty(t, res.GQL.Errors)
	subCh := res.Subscription
	require.NotNil(t, subCh)

	authorDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name": "Tolkien"}`), authorCol.Version())
	require.NoError(t, err)
	require.NoError(t, authorCol.AddDocument(ctx, authorDoc))

	// Capture the txn counter AFTER the mutation's own work returned.
	// The subscription event handler runs on its own goroutine; we then
	// wait and assert the counter has not advanced further.
	mid := db.previousTxnID.Load()
	time.Sleep(200 * time.Millisecond)
	after := db.previousTxnID.Load()

	require.Equal(t, mid, after,
		"subscription must not open a transaction for a wrong-collection event; counter advanced from %d to %d",
		mid, after)

	// Belt-and-braces: also confirm nothing surfaces on the response channel.
	select {
	case got, ok := <-subCh:
		if !ok {
			t.Fatalf("subscription channel closed unexpectedly")
		}
		t.Fatalf("expected no response for wrong-collection event, got %+v", got)
	default:
	}
}

// Control: same-collection events must still be delivered. Guards against the
// new collection filter being over-aggressive.
func TestHandleSubscription_RightCollectionEvent_StillDelivered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, bookAuthorSchema)
	require.NoError(t, err)

	bookCol, err := db.GetCollectionByName(ctx, "Book")
	require.NoError(t, err)

	res := db.ExecRequest(ctx, `subscription { Book { _docID title } }`)
	require.Empty(t, res.GQL.Errors)
	subCh := res.Subscription
	require.NotNil(t, subCh)

	bookDoc, err := client.NewDocFromJSON(ctx, []byte(`{"title": "The Hobbit"}`), bookCol.Version())
	require.NoError(t, err)
	require.NoError(t, bookCol.AddDocument(ctx, bookDoc))

	select {
	case got, ok := <-subCh:
		require.True(t, ok, "subscription channel must stay open")
		require.Empty(t, got.Errors, "right-collection event must deliver without errors")
		require.NotEmpty(t, got.Data)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected right-collection event to be delivered")
	}
}
