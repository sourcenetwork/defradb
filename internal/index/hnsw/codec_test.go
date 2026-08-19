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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Node round-trip: field preservation across the representative shapes ---

func TestMarshalNode_RepresentativeShapes_RoundTrip(t *testing.T) {
	cases := map[string]Node{
		"all fields with an empty-but-present layer": {
			ID:         42,
			Vector:     []float32{0.1, -0.2, 0.3, 1.5, -9.75},
			Neighbours: [][]NodeID{{1, 2, 3}, {4}, {}, {5, 6, 7, 8, 9}},
			Deleted:    true,
		},
		"present-but-empty vector and nil layers": {
			ID: 7, Vector: []float32{}, Neighbours: nil,
		},
		"zero-value node (nil vector, nil layers)": {
			ID: 3,
		},
		"single element vector, single layer": {
			ID: 8, Vector: []float32{1.25}, Neighbours: [][]NodeID{{9}},
		},
	}

	for name, n := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := MarshalNode(n)
			require.NoError(t, err)
			out, err := UnmarshalNode(b)
			require.NoError(t, err)
			// Empty/nil slices are canonicalised to nil on decode; compare via the accessors that
			// matter rather than requiring []float32{} == nil.
			assert.Equal(t, n.ID, out.ID)
			assert.Equal(t, n.Deleted, out.Deleted)
			assert.Equal(t, len(n.Vector), len(out.Vector))
			for i := range n.Vector {
				assert.Equal(t, n.Vector[i], out.Vector[i])
			}
			assert.Equal(t, len(n.Neighbours), len(out.Neighbours))
			for i := range n.Neighbours {
				assert.Equal(t, n.Neighbours[i], out.Neighbours[i])
			}
		})
	}
}

func TestMarshalNode_Deterministic_SameBytes(t *testing.T) {
	n := Node{ID: 123, Vector: []float32{1, 2, 3}, Neighbours: [][]NodeID{{1, 2}, {3}}}

	b1, err := MarshalNode(n)
	require.NoError(t, err)
	b2, err := MarshalNode(n)
	require.NoError(t, err)
	assert.Equal(t, b1, b2)
}

// --- Node decode failure modes ---

// TestUnmarshalNode_CorruptInput_ReturnsError covers the fixed corruptions that each target a
// distinct decode guard. Truncation (which sweeps a range of lengths) is a separate test below.
func TestUnmarshalNode_CorruptInput_ReturnsError(t *testing.T) {
	valid := func(t *testing.T) []byte {
		t.Helper()
		b, err := MarshalNode(Node{ID: 1, Vector: []float32{1, 2}, Neighbours: [][]NodeID{{9}}})
		require.NoError(t, err)
		return b
	}

	cases := map[string]func(t *testing.T) []byte{
		"unsupported version byte": func(t *testing.T) []byte {
			b := valid(t)
			b[0] = 0xFF
			return b
		},
		"trailing bytes past a valid encoding": func(t *testing.T) []byte {
			return append(valid(t), 0x00)
		},
		"vector-length prefix larger than the buffer": func(t *testing.T) []byte {
			b := valid(t)
			off := versionWidth + nodeIDWidth + flagWidth
			b[off], b[off+1], b[off+2], b[off+3] = 0xFF, 0xFF, 0xFF, 0x0F
			return b
		},
	}

	for name, makeInput := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := UnmarshalNode(makeInput(t))
			assert.ErrorIs(t, err, ErrInvalidNodeEncoding)
			assert.Equal(t, Node{}, out)
		})
	}
}

func TestUnmarshalNode_TruncatedAtEveryBoundary_ReturnsError(t *testing.T) {
	// A node whose encoding exercises each internal bounds check (vector-length prefix, vector
	// body, layer-count prefix, neighbour-count prefix, neighbour body). Truncating at every
	// length short of the full buffer must error at some bounds check and never panic.
	b, err := MarshalNode(Node{ID: 1, Vector: []float32{1, 2, 3}, Neighbours: [][]NodeID{{1, 2, 3}, {4}}})
	require.NoError(t, err)

	for truncLen := range len(b) {
		out, err := UnmarshalNode(b[:truncLen])
		assert.ErrorIs(t, err, ErrInvalidNodeEncoding, "truncation length %d should error", truncLen)
		assert.Equal(t, Node{}, out)
	}
}

// --- Meta round-trip and failure modes ---

func TestMarshalMeta_RepresentativeValues_RoundTrip(t *testing.T) {
	cases := map[string]Meta{
		"populated":          {EntryPoint: 99, TopLayer: 4},
		"zero value":         {EntryPoint: 0, TopLayer: 0},
		"negative top layer": {EntryPoint: 123456789, TopLayer: -1},
		"large entry point":  {EntryPoint: 1 << 40, TopLayer: 7},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := MarshalMeta(m)
			require.NoError(t, err)
			out, err := UnmarshalMeta(b)
			require.NoError(t, err)
			assert.Equal(t, m, out)
		})
	}
}

func TestMarshalMeta_Deterministic_SameBytes(t *testing.T) {
	m := Meta{EntryPoint: 77, TopLayer: 3}

	b1, err := MarshalMeta(m)
	require.NoError(t, err)
	b2, err := MarshalMeta(m)
	require.NoError(t, err)
	assert.Equal(t, b1, b2)
}

func TestUnmarshalMeta_CorruptInput_ReturnsError(t *testing.T) {
	valid := func(t *testing.T) []byte {
		t.Helper()
		b, err := MarshalMeta(Meta{EntryPoint: 5, TopLayer: 2})
		require.NoError(t, err)
		return b
	}

	cases := map[string]func(t *testing.T) []byte{
		"unsupported version byte": func(t *testing.T) []byte {
			b := valid(t)
			b[0] = 0xFF
			return b
		},
		"too short": func(t *testing.T) []byte {
			b := valid(t)
			return b[:len(b)-1]
		},
		"too long": func(t *testing.T) []byte {
			return append(valid(t), 0x00)
		},
	}

	for name, makeInput := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := UnmarshalMeta(makeInput(t))
			assert.ErrorIs(t, err, ErrInvalidMetaEncoding)
			assert.Equal(t, Meta{}, out)
		})
	}
}

// --- Integration: real engine-produced nodes/meta survive the codec ---

// TestCodec_EngineProducedGraph_RoundTripsThroughBytes builds a graph via the standard engine
// entry points and verifies that the actual node shapes and Meta the engine emits (varying vector
// lengths, multi-layer adjacency, a real entry point) survive marshal/unmarshal. This complements
// the hand-crafted unit cases above by covering shapes the engine actually produces.
func TestCodec_EngineProducedGraph_RoundTripsThroughBytes(t *testing.T) {
	const (
		n   = 50
		dim = 12
	)

	store := NewMemStore()
	g := New(store, Cosine, DefaultParams(8), 11)

	rng := rand.New(rand.NewSource(11))
	for i := range n {
		v := make([]float32, dim)
		for j := range v {
			v[j] = float32(rng.NormFloat64())
		}
		require.NoError(t, g.Insert(NodeID(i+1), v))
	}

	sawMultiLayer := false
	err := store.IterateNodes(func(original Node) error {
		b, marshalErr := MarshalNode(original)
		if marshalErr != nil {
			return marshalErr
		}
		out, unmarshalErr := UnmarshalNode(b)
		if unmarshalErr != nil {
			return unmarshalErr
		}
		assert.Equal(t, original, out)
		if len(original.Neighbours) > 1 {
			sawMultiLayer = true
		}
		return nil
	})
	require.NoError(t, err)
	assert.True(t, sawMultiLayer, "engine should produce at least one multi-layer node to exercise layer encoding")

	meta, _, err := store.GetMeta()
	require.NoError(t, err)
	b, err := MarshalMeta(meta)
	require.NoError(t, err)
	outMeta, err := UnmarshalMeta(b)
	require.NoError(t, err)
	assert.Equal(t, meta, outMeta)
}
