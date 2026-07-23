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

// Search returns up to k of the nearest non-deleted node ids to query,
// nearest-first, following Algorithm 5 (K-NN-SEARCH) of the HNSW paper:
// descend greedily (ef=1) from the entry point at the top layer down to
// layer 1, then run SEARCH-LAYER at layer 0 with ef=max(efSearch, k).
//
// If the graph is empty, Search returns an empty (nil) slice and no
// error.
func (g *Graph) Search(query []float32, k, efSearch int) ([]NodeID, error) {
	if k <= 0 {
		return nil, nil
	}
	if efSearch < k {
		efSearch = k
	}

	meta, err := g.store.GetMeta()
	if err != nil {
		return nil, err
	}
	if meta.Empty {
		return nil, nil
	}

	q := normalize(query)

	entry, ok, err := g.store.GetNode(meta.EntryPoint)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	curBest := candidate{id: entry.ID, dist: distance(g.metric, q, entry.Vector), vector: entry.Vector}

	for layer := meta.TopLayer; layer > 0; layer-- {
		w, err := g.searchGreedy(q, curBest.id, layer)
		if err != nil {
			return nil, err
		}
		if len(w) > 0 {
			curBest = w[0]
		}
	}

	w, err := g.searchLayerMulti(q, []candidate{curBest}, efSearch, 0)
	if err != nil {
		return nil, err
	}

	if len(w) > k {
		w = w[:k]
	}

	out := make([]NodeID, len(w))
	for i, c := range w {
		out[i] = c.id
	}
	return out, nil
}

// Delete tombstones the node with the given id: it marks the node as
// deleted but leaves its links intact so that graph connectivity (and
// traversal through it) is preserved. Deleted nodes are excluded from
// Search results.
//
// TODO: if the entry point itself is deleted, search still works because
// traversal passes through deleted nodes, but a full reclaim/repair pass
// (e.g. picking a new entry point, unlinking tombstones) is left for a
// later phase.
func (g *Graph) Delete(id NodeID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, ok, err := g.store.GetNode(id)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if node.Deleted {
		return nil
	}
	node.Deleted = true
	return g.store.PutNode(node)
}
