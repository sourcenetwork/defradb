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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// TestCollectionAddManyDocuments_EmptyBatch_IsNoOp test that AddManyDocuments's JSON bridge properly
// handles a nil/empty docs slice. jsonDocs is built via make([]json.RawMessage, 0, len(docs)),
// which is non-nil even when docs is empty, so json.Marshal emits "[]" rather than "null."
// AddDocumentNative's leading '[' check (see: cbindings/document_add.go) then classifies this as a
// (valid, empty) batch rather than failing to parse a single document out of "null". The Go client's 
// AddManyDocuments treats an empty batch as a successful no-op, so this must do that too.
func TestCollectionAddManyDocuments_EmptyBatch_IsNoOp(t *testing.T) {
	w, ctx := newTestWrapper(t)

	cols, err := w.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)
	require.Len(t, cols, 1)

	// A nil slice
	t.Run("nil docs", func(t *testing.T) {
		require.NoError(t, cols[0].AddManyDocuments(ctx, nil))
	})

	// An explicitly empty, non-nil slice
	t.Run("empty docs", func(t *testing.T) {
		require.NoError(t, cols[0].AddManyDocuments(ctx, []*client.Document{}))
	})
}

// TestCollectionSaveDocument_AddAndUpdatePaths_Succeed is a functional regression test for
// SaveDocument wrapping its GetDocument/Add-or-UpdateDocument sequence in its own transaction
// (when the caller hasn't already attached one) to close the race window between the two calls.
func TestCollectionSaveDocument_AddAndUpdatePaths_Succeed(t *testing.T) {
	w, ctx := newTestWrapper(t)

	cols, err := w.GetCollections(ctx, options.GetCollections().SetCollectionName("Users"))
	require.NoError(t, err)
	require.Len(t, cols, 1)
	col := cols[0]

	doc, err := client.NewDocFromMap(ctx, map[string]any{"name": "Alice"}, col.Version())
	require.NoError(t, err)

	// Add path: the document doesn't exist yet, so SaveDocument's internal GetDocument must
	// return not-found and it must fall through to AddDocument.
	require.NoError(t, col.SaveDocument(ctx, doc))

	fetched, err := col.GetDocument(ctx, doc.ID(), options.GetDocument())
	require.NoError(t, err)
	name, err := fetched.Get("name")
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	// Update path: the document now exists, so SaveDocument's internal GetDocument must succeed
	// and it must fall through to UpdateDocument instead of failing with a duplicate-key error.
	require.NoError(t, doc.Set(ctx, "name", "Bob"))
	require.NoError(t, col.SaveDocument(ctx, doc))

	fetched, err = col.GetDocument(ctx, doc.ID(), options.GetDocument())
	require.NoError(t, err)
	name, err = fetched.Get("name")
	require.NoError(t, err)
	require.Equal(t, "Bob", name)
}