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

// Neighbor is one search hit: a node id and its distance to the query vector (smaller is nearer).
type Neighbor struct {
	ID       NodeID
	Distance float64
}

// Search returns up to k of the nearest non-deleted node ids to query,
// nearest-first, following Algorithm 5 (K-NN-SEARCH) of the HNSW paper:
// descend greedily (ef=1) from the entry point at the top layer down to
// layer 1, then run SEARCH-LAYER at layer 0 with ef=max(efSearch, k).
//
// If the graph is empty, Search returns an empty (nil) slice and no
// error.
func (g *HNSWIndex) Search(query []float32, k, efSearch int) ([]NodeID, error) {
	w, err := g.search(query, k, efSearch)
	if err != nil {
		return nil, err
	}
	out := make([]NodeID, len(w))
	for i, c := range w {
		out[i] = c.id
	}
	return out, nil
}

// SearchWithDistance is Search but each hit also carries its distance to the query. The caller uses
// the distance to rank or score results (e.g. to fill a similarity field) without recomputing it.
func (g *HNSWIndex) SearchWithDistance(query []float32, k, efSearch int) ([]Neighbor, error) {
	w, err := g.search(query, k, efSearch)
	if err != nil {
		return nil, err
	}
	out := make([]Neighbor, len(w))
	for i, c := range w {
		out[i] = Neighbor{ID: c.id, Distance: float64(c.dist)}
	}
	return out, nil
}

// search runs the K-NN-SEARCH traversal and returns the top-k candidates (with distances),
// nearest-first. Both Search and SearchWithDistance are thin wrappers over this.
func (g *HNSWIndex) search(query []float32, k, efSearch int) ([]candidate, error) {
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
	return w, nil
}

// Delete tombstones the node: it stays linked so traversal through it is preserved, but is excluded
// from Search results. Links are left intact deliberately, even when the entry point is deleted,
// because traversal still passes through tombstones. Physically unlinking them and promoting a new
// entry point needs a reclaim pass that is not yet built.
func (g *HNSWIndex) Delete(id NodeID) error {
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
