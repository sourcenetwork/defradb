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

package java

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// TestTxnCommit_ConflictingWrite_ReturnsError forces a real native commit failure by having two 
// independent transactions read-then-write the same document. The first Commit should win and 
// succeed. The second should fail with a conflict.
func TestTxnCommit_ConflictingWrite_ReturnsError(t *testing.T) {
	w, ctx := newTestWrapper(t)

	cols, err := w.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)
	require.Len(t, cols, 1)

	doc, err := client.NewDocFromMap(ctx, map[string]any{"name": "Alice"}, cols[0].Version())
	require.NoError(t, err)
	require.NoError(t, cols[0].AddDocument(ctx, doc))

	txn1, err := w.NewTxn(false)
	require.NoError(t, err)
	defer txn1.Discard()
	txn2, err := w.NewTxn(false)
	require.NoError(t, err)
	defer txn2.Discard()

	cols1, err := txn1.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)
	cols2, err := txn2.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)

	doc1, err := cols1[0].GetDocument(ctx, doc.ID(), options.GetDocument())
	require.NoError(t, err)
	require.NoError(t, doc1.Set(ctx, "name", "Bob"))
	require.NoError(t, cols1[0].UpdateDocument(ctx, doc1))

	doc2, err := cols2[0].GetDocument(ctx, doc.ID(), options.GetDocument())
	require.NoError(t, err)
	require.NoError(t, doc2.Set(ctx, "name", "Carol"))
	require.NoError(t, cols2[0].UpdateDocument(ctx, doc2))

	require.NoError(t, txn1.Commit())

	err = txn2.Commit()
	require.Error(t, err, "txn2's commit should fail: it conflicts with txn1's already-committed write")

	require.ErrorIs(t, txn2.Commit(), client.ErrTransactionNotFound)
}

// TestTxn_RepeatedCommitAndDiscard checks every combination of a second finalization call. It
// must never repeat the native commit/discard call (which would panic on the already-deleted
// handle), and must report the outcome honestly rather than defaulting to nil/success.
func TestTxn_RepeatedCommitAndDiscard(t *testing.T) {
	w, ctx := newTestWrapper(t)

	newTxn := func(t *testing.T) client.Txn {
		t.Helper()
		txn, err := w.NewTxn(false)
		require.NoError(t, err)
		return txn
	}
	_ = ctx

	t.Run("commit twice", func(t *testing.T) {
		txn := newTxn(t)
		require.NoError(t, txn.Commit())
		require.ErrorIs(t, txn.Commit(), client.ErrTransactionNotFound)
	})

	t.Run("discard after commit is a no-op", func(t *testing.T) {
		txn := newTxn(t)
		require.NoError(t, txn.Commit())
		require.NotPanics(t, txn.Discard)
	})

	t.Run("commit after discard returns an error", func(t *testing.T) {
		txn := newTxn(t)
		txn.Discard()
		require.ErrorIs(t, txn.Commit(), client.ErrTransactionNotFound)
	})

	t.Run("discard twice is a no-op", func(t *testing.T) {
		txn := newTxn(t)
		txn.Discard()
		require.NotPanics(t, txn.Discard)
	})
}

// TestTxnPostFinalization_Operations_ReturnTransactionNotFound guards callStore's finalized
// check. A Store/Collection call made through a context still carrying an already-finalized
// transaction must fail with client.ErrTransactionNotFound instead of reusing the transaction's
// now-zeroed handle/txnObj.
func TestTxnPostFinalization_Operations_ReturnTransactionNotFound(t *testing.T) {
	w, ctx := newTestWrapper(t)

	txn, err := w.NewTxn(false)
	require.NoError(t, err)
	require.NoError(t, txn.Commit())

	_, err = txn.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.ErrorIs(t, err, client.ErrTransactionNotFound)

	_, err = txn.AddCollection(ctx, `type Other { value: String }`)
	require.ErrorIs(t, err, client.ErrTransactionNotFound)
}

// TestTxnCommitDiscard_Concurrent_NoRace guards finalizeMu. Commit and Discard racing each other
// on the same transaction must never both reach the native layer, which would double-delete the
// same cgo.Handle/JNI global reference. 
// 
// Run with -race. 
// Also run with DEFRA_JAVA_JVM_OPTS=-Xcheck:jni to make the JVM itself abort on invalid reference
// use instead of silently tolerating it.
func TestTxnCommitDiscard_Concurrent_NoRace(t *testing.T) {
	w, _ := newTestWrapper(t)

	const trials = 20
	for i := 0; i < trials; i++ {
		txn, err := w.NewTxn(false)
		require.NoError(t, err)

		var wg sync.WaitGroup
		var commitErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			commitErr = txn.Commit()
		}()
		go func() {
			defer wg.Done()
			txn.Discard()
		}()
		wg.Wait()

		// Whichever of Commit/Discard actually ran first, the transaction must end up finalized.
		require.ErrorIs(t, txn.Commit(), client.ErrTransactionNotFound)
		require.NotPanics(t, txn.Discard)
		t.Logf("trial %d: commitErr=%v", i, commitErr)
	}
}