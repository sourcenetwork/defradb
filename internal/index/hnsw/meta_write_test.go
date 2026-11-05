// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package hnsw

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// putMetaCounter wraps a NodeStore and counts PutMeta calls.
type putMetaCounter struct {
	NodeStore
	puts int
}

func (s *putMetaCounter) PutMeta(m Meta) error {
	s.puts++
	return s.NodeStore.PutMeta(m)
}

// Every insert reads the meta singleton, so writing it when nothing changed makes all concurrent
// inserts conflict under optimistic transactions. Insert must only write meta when it changes.
func TestInsert_WritesMetaOnlyWhenChanged(t *testing.T) {
	store := &putMetaCounter{NodeStore: NewMemStore()}
	g := New(store, Cosine, DefaultParams(8), 1)

	// ML == 0 makes randomLevel always return 0; a huge ML always returns a very high level.
	g.params.ML = 0

	require.NoError(t, g.Insert(1, []float32{1, 0, 0}))
	require.Equal(t, 1, store.puts, "first insert must write meta")

	require.NoError(t, g.Insert(2, []float32{0, 1, 0}))
	require.NoError(t, g.Insert(3, []float32{0, 0, 1}))
	require.Equal(t, 1, store.puts, "inserts that keep entry point and top layer must not write meta")

	g.params.ML = 1e9
	require.NoError(t, g.Insert(4, []float32{1, 1, 0}))
	require.Equal(t, 2, store.puts, "insert that raises the top layer must write meta")

	meta, _, err := store.GetMeta()
	require.NoError(t, err)
	require.Equal(t, NodeID(4), meta.EntryPoint)
	require.Greater(t, meta.TopLayer, 0)
}
