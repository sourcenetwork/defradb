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

// Node is the persisted form of a graph node: its (normalized) vector plus
// per-layer adjacency.
type Node struct {
	ID     NodeID
	Vector []float32 // normalized
	// Neighbours[l] is this node's own neighbour ids at layer l (node-scoped adjacency, not the
	// global layer l). len(Neighbours) = node's top level + 1.
	Neighbours [][]NodeID
	Deleted    bool // tombstone
}

// Meta is the graph's global state: which node to start a search from, and the top layer.
type Meta struct {
	EntryPoint NodeID
	TopLayer   int
}

// NodeStore is the persistence port. The engine never imports a concrete
// store; implementations may back this interface with any storage medium
// (in-memory, a KV store, etc).
type NodeStore interface {
	// GetNode returns the node with the given id. found is false when no such node has been stored,
	// in which case Node is the zero value.
	GetNode(id NodeID) (n Node, found bool, err error)
	// PutNode stores n, replacing any previously stored node with the same id.
	PutNode(n Node) error
	// GetMeta returns the graph's meta singleton. found is false when no graph has been built yet
	// (nothing stored), in which case Meta is the zero value and there is no entry point to start from.
	GetMeta() (m Meta, found bool, err error)
	// PutMeta stores m as the graph's meta singleton, replacing any previously stored value.
	PutMeta(m Meta) error
	// IterateNodes iterates over all non-deleted node ids (used for
	// brute-force baselines / rebuilds). Order is unspecified. Iteration
	// stops early if fn returns an error, and that error is returned to
	// the caller of IterateNodes.
	IterateNodes(fn func(Node) error) error
}
