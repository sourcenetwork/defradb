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