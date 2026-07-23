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

import "sync"

// Node is the persisted form of a graph node: its (normalized) vector plus
// per-layer adjacency.
type Node struct {
	ID      NodeID
	Vector  []float32  // normalized
	Layers  [][]NodeID // Layers[l] = neighbour ids at layer l; len(Layers) = node's top level + 1
	Deleted bool       // tombstone
}

// Meta is the graph's global state.
type Meta struct {
	EntryPoint NodeID
	TopLayer   int
	Empty      bool // true when no entry point has been set yet
}

// NodeStore is the persistence port. The engine never imports a concrete
// store; implementations may back this interface with any storage medium
// (in-memory, a KV store, etc).
type NodeStore interface {
	GetNode(id NodeID) (Node, bool, error)
	PutNode(n Node) error
	GetMeta() (Meta, error)
	PutMeta(m Meta) error
	// IterateNodes iterates over all non-deleted node ids (used for
	// brute-force baselines / rebuilds). Order is unspecified. Iteration
	// stops early if fn returns an error, and that error is returned to
	// the caller of IterateNodes.
	IterateNodes(fn func(Node) error) error
}

// memStore is an in-memory NodeStore implementation, safe for concurrent
// use. It is useful for tests and for standalone/in-RAM use of the graph.
type memStore struct {
	mu    sync.Mutex
	nodes map[NodeID]Node
	meta  Meta
}

// NewMemStore returns a new in-memory NodeStore.
func NewMemStore() NodeStore {
	return &memStore{
		nodes: make(map[NodeID]Node),
		meta:  Meta{Empty: true},
	}
}

func (s *memStore) GetNode(id NodeID) (Node, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.nodes[id]
	if !ok {
		return Node{}, false, nil
	}
	return cloneNode(n), true, nil
}

func (s *memStore) PutNode(n Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[n.ID] = cloneNode(n)
	return nil
}

func (s *memStore) GetMeta() (Meta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.meta, nil
}

func (s *memStore) PutMeta(m Meta) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.meta = m
	return nil
}

func (s *memStore) IterateNodes(fn func(Node) error) error {
	// Snapshot the node list under lock, then invoke fn outside the lock
	// so that fn may itself call back into the store without deadlocking.
	s.mu.Lock()
	nodes := make([]Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		if n.Deleted {
			continue
		}
		nodes = append(nodes, cloneNode(n))
	}
	s.mu.Unlock()

	for _, n := range nodes {
		if err := fn(n); err != nil {
			return err
		}
	}
	return nil
}

// cloneNode returns a deep copy of n so that callers cannot mutate the
// store's internal state through returned/stored slices.
func cloneNode(n Node) Node {
	out := Node{
		ID:      n.ID,
		Deleted: n.Deleted,
	}
	if n.Vector != nil {
		out.Vector = make([]float32, len(n.Vector))
		copy(out.Vector, n.Vector)
	}
	if n.Layers != nil {
		out.Layers = make([][]NodeID, len(n.Layers))
		for i, layer := range n.Layers {
			if layer == nil {
				continue
			}
			l := make([]NodeID, len(layer))
			copy(l, layer)
			out.Layers[i] = l
		}
	}
	return out
}
