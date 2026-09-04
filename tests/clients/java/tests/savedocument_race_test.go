// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build javaclient

package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// TestCollectionSaveDocument_RacesDeleteDocument_PostFix is a diagnostic test mirroring the one
// used against the (still un-fixed) C bindings client, run here against the java client's
// now-transaction-wrapped SaveDocument to check what the fix actually closes.
//
// Each trial races SaveDocument (update path) against a concurrent, independent DeleteDocument
// call on the same document. Two distinct failure modes are checked for separately:
//   - the wrong-error race: SaveDocument, having just confirmed the document exists via its
//     internal GetDocument, fails only because the concurrent delete won - this is what the
//     transaction-wrapping fix targets, since both of SaveDocument's own steps now run inside one
//     transaction and should be arbitrated by DefraDB's normal transaction-conflict handling
//     instead of racing the delete's separate transaction directly.
//   - the deadlock: the trial not completing within a bounded time at all. Wrapping SaveDocument's
//     own steps in a transaction doesn't change that a concurrent DeleteDocument is still a
//     genuinely separate, independently-scheduled transaction hitting the same in-memory store -
//     the earlier deadlock found on C was in the datastore's own locking under concurrent
//     transactions, not specific to how many high-level calls SaveDocument happens to make, so
//     there's no a priori reason to expect this fix to prevent it.
func TestCollectionSaveDocument_RacesDeleteDocument_PostFix(t *testing.T) {
	w, ctx := newTestWrapper(t)

	cols, err := w.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)
	require.Len(t, cols, 1)
	col := cols[0]

	const trials = 20
	wrongErrorReproduced := false
	for i := 0; i < trials; i++ {
		doc, err := client.NewDocFromMap(ctx, map[string]any{"name": fmt.Sprintf("Alice-%d", i)}, col.Version())
		require.NoError(t, err)
		require.NoError(t, col.AddDocument(ctx, doc))

		updateDoc, err := client.NewDocWithID(ctx, doc.ID(), col.Version())
		require.NoError(t, err)
		require.NoError(t, updateDoc.Set(ctx, "name", fmt.Sprintf("Bob-%d", i)))

		var wg sync.WaitGroup
		var saveErr, delErr error
		start := make(chan struct{})
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			saveErr = col.SaveDocument(ctx, updateDoc)
		}()
		go func() {
			defer wg.Done()
			<-start
			_, delErr = col.DeleteDocument(ctx, doc.ID())
		}()
		close(start)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			t.Logf("trial %d: saveErr=%v delErr=%v", i, saveErr, delErr)
			if delErr == nil && saveErr != nil {
				wrongErrorReproduced = true
				t.Logf("trial %d: wrong-error race reproduced despite the transaction-wrapping fix: %v", i, saveErr)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("trial %d: DEADLOCK - SaveDocument/DeleteDocument race did not complete within 5s", i)
		}
	}

	t.Logf("wrong-error race reproduced in at least one of %d trials: %v", trials, wrongErrorReproduced)
}
