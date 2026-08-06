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
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// HNSWIndex is a Hierarchical Navigable Small World approximate-nearest-neighbour index. It is safe
// for concurrent use: mutations (Insert, Delete) are serialized via an internal mutex (single-writer
// discipline), while Search may run concurrently with mutations as long as the underlying NodeStore
// provides safe/consistent concurrent reads (the provided in-memory store guards all access with its
// own mutex; a real KV-backed store would typically offer snapshot reads).
type HNSWIndex struct {
	store  NodeStore
	params Params
	metric Metric

	rng *rand.Rand
	mu  sync.Mutex // serializes Insert/Delete (single-writer)
}

// New creates a new HNSWIndex backed by the given store, using the given
// distance metric and construction/search parameters. seed makes level
// generation deterministic for a given sequence of inserts, which is
// useful for tests and reproducibility.
func New(store NodeStore, metric Metric, params Params, seed int64) *HNSWIndex {
	return &HNSWIndex{
		store:  store,
		params: params,
		metric: metric,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// randomLevel draws a random top layer for a new node, following the exponential-decay level
// distribution from the HNSW paper: l = floor(-ln(unif(0,1)) * ML).
//
// The decay is what makes the graph a hierarchy: almost every node lands on the bottom layer, and
// each layer up holds exponentially fewer nodes. Search starts at the sparse top and drops down, so
// it can cover a lot of ground in a few steps before doing the fine-grained search at the bottom.
// ML controls how fast the layers thin out.
func (g *HNSWIndex) randomLevel() int {
	// rng.Float64 returns a value in [0, 1); guard against exactly 0 to
	// avoid taking ln(0).
	r := g.rng.Float64()
	for r == 0 {
		r = g.rng.Float64()
	}
	return int(math.Floor(-math.Log(r) * g.params.ML))
}

// Insert adds a new vector under the given node id, following Algorithm 1
// of the HNSW paper (INSERT).
func (g *HNSWIndex) Insert(id NodeID, vector []float32) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	v := normalize(vector)
	topLevel := g.randomLevel()

	meta, found, err := g.store.GetMeta()
	if err != nil {
		return err
	}

	if !found {
		layers := make([][]NodeID, topLevel+1)
		for i := range layers {
			layers[i] = []NodeID{}
		}
		if err := g.store.PutNode(Node{ID: id, Vector: v, Layers: layers}); err != nil {
			return err
		}
		return g.store.PutMeta(Meta{EntryPoint: id, TopLayer: topLevel})
	}

	// Descent: greedily walk down from the entry point at the top layer
	// to layer topLevel+1, keeping only the single closest node found so
	// far (ef=1 search at each layer).
	entry, found, err := g.store.GetNode(meta.EntryPoint)
	if err != nil {
		return err
	}
	if !found {
		// meta references a node the store does not have, so the graph is torn. Insert cannot descend
		// from a missing entry point, so fail rather than build links from a zero-value node.
		return ErrEntryPointNotFound
	}
	curBest := candidate{id: entry.ID, dist: distance(g.metric, v, entry.Vector), vector: entry.Vector}

	for layer := meta.TopLayer; layer > topLevel; layer-- {
		w, err := g.searchGreedy(v, curBest.id, layer)
		if err != nil {
			return err
		}
		if len(w) > 0 {
			curBest = w[0]
		}
	}

	// Insertion: from min(topLevel, meta.TopLayer) down to 0, search with
	// ef=efConstruction, select diverse neighbours, and link.
	newLayers := make([][]NodeID, topLevel+1)
	for i := range newLayers {
		newLayers[i] = []NodeID{}
	}

	startLayer := min(meta.TopLayer, topLevel)

	entryPoints := []candidate{curBest}
	for layer := startLayer; layer >= 0; layer-- {
		w, err := g.searchLayerMulti(v, entryPoints, g.params.EfConstruction, layer)
		if err != nil {
			return err
		}

		mmax := g.params.M
		if layer == 0 {
			mmax = g.params.Mmax0
		}

		selected := selectNeighborsHeuristic(g.metric, w, g.params.M)
		newLayers[layer] = idsOf(selected)

		// Add bidirectional links and shrink neighbours that exceed mmax.
		for _, sel := range selected {
			if err := g.addLink(sel.id, id, layer, mmax); err != nil {
				return err
			}
		}

		// Start the next layer down from every neighbour found here, not just the closest. A node
		// exists on every layer up to its own height, so these are all valid entry points below.
		entryPoints = w
	}

	if err := g.store.PutNode(Node{ID: id, Vector: v, Layers: newLayers}); err != nil {
		return err
	}

	if topLevel > meta.TopLayer {
		meta.EntryPoint = id
		meta.TopLayer = topLevel
	}

	// If every node reachable from the entry point is tombstoned (e.g. all documents were deleted and
	// new ones inserted), the search above returns no live neighbours and this node links to nothing.
	// It would then be unreachable from the entry point. Promote it to entry point so search can find
	// it and later inserts can attach to it.
	if len(newLayers[0]) == 0 {
		meta.EntryPoint = id
		if topLevel > meta.TopLayer {
			meta.TopLayer = topLevel
		}
	}

	return g.store.PutMeta(meta)
}

// addLink adds a bidirectional edge between "from" and "to" at the given
// layer: it appends "to" to from's neighbour list at that layer, then, if
// from's degree at that layer now exceeds mmax, re-runs the neighbour
// selection heuristic over from's neighbours and prunes to the cap.
func (g *HNSWIndex) addLink(from, to NodeID, layer, mmax int) error {
	node, ok, err := g.store.GetNode(from)
	if err != nil {
		return err
	}
	if !ok {
		// A backlink can name a node that a later reclaim pass removed. Skip the missing node rather
		// than error: it just means one fewer back-edge, which search tolerates.
		return nil
	}

	// A node's top layer is chosen at random on insert, so "from" may have been created with fewer
	// layers than the one we are linking at. Grow it with empty neighbour lists up to that layer; the
	// empty lists are valid (no neighbours yet), and the link is added on the last one below.
	for len(node.Layers) <= layer {
		node.Layers = append(node.Layers, []NodeID{})
	}

	node.Layers[layer] = append(node.Layers[layer], to)

	if len(node.Layers[layer]) > mmax {
		candidates, err := g.candidatesFromIDs(node.Vector, node.Layers[layer])
		if err != nil {
			return err
		}
		selected := selectNeighborsHeuristic(g.metric, candidates, mmax)
		node.Layers[layer] = idsOf(selected)
	}

	return g.store.PutNode(node)
}

// candidatesFromIDs loads the nodes for the given ids and computes their
// distance to q, returning them as candidates (unsorted).
func (g *HNSWIndex) candidatesFromIDs(q []float32, ids []NodeID) ([]candidate, error) {
	out := make([]candidate, 0, len(ids))
	for _, id := range ids {
		n, ok, err := g.store.GetNode(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			// ids come from a neighbour list, which can name a node a reclaim pass removed. Skip it.
			continue
		}
		out = append(out, candidate{id: id, dist: distance(g.metric, q, n.Vector), vector: n.Vector})
	}
	return out, nil
}

// idsOf extracts the node ids from a candidate slice.
func idsOf(cands []candidate) []NodeID {
	ids := make([]NodeID, len(cands))
	for i, c := range cands {
		ids[i] = c.id
	}
	return ids
}

// searchGreedy runs the ef=1 descent step of SEARCH-LAYER (Algorithm 2) from a single entry point:
// it returns the single closest node reachable in the given layer. This is the per-layer step used
// when descending from the top layer toward the insertion/query layer.
func (g *HNSWIndex) searchGreedy(query []float32, entry NodeID, layer int) ([]candidate, error) {
	return g.searchLayerMulti(query, []candidate{{id: entry}}, 1, layer)
}

// searchLayerMulti runs SEARCH-LAYER (Algorithm 2), the ef-bounded greedy search of a single layer,
// starting from a set of entry points. It takes a set, not one point, because Insert seeds each layer
// with all the neighbours found on the layer above; the descent step passes a single point.
// Distances on the supplied entryPoints are recomputed against query as needed.
//
// Deleted (tombstoned) nodes are traversed (their neighbours are still
// explored, keeping the graph connected) but are never added to the
// result set W, per the package's tombstone-deletion strategy.
//
// The returned candidates are sorted nearest-first.
func (g *HNSWIndex) searchLayerMulti(query []float32, entryPoints []candidate, ef, layer int) ([]candidate, error) {
	visited := make(map[NodeID]struct{})
	candidatesHeap := newMinHeap()
	results := newMaxHeap()

	for _, ep := range entryPoints {
		n, ok, err := g.store.GetNode(ep.id)
		if err != nil {
			return nil, err
		}
		if !ok {
			// An entry point can name a node a reclaim pass removed. Skip it and try the next.
			continue
		}
		// Today's callers pass unique entry points, so this never fires. It is a cheap guard that keeps
		// the walk correct for any input: a duplicate entry point would otherwise be counted twice.
		if _, seen := visited[ep.id]; seen {
			continue
		}
		visited[ep.id] = struct{}{}

		d := distance(g.metric, query, n.Vector)
		c := candidate{id: ep.id, dist: d, vector: n.Vector}
		candidatesHeap.items = append(candidatesHeap.items, c)
		if !n.Deleted {
			results.items = append(results.items, c)
		}
	}
	heap.Init(candidatesHeap)
	heap.Init(results)

	// Greedy walk from the entry points: follow edges outward, never scanning the whole layer. The
	// entry points decide where the walk starts and which part of the graph it reaches, so better ones
	// give better results.
	for candidatesHeap.Len() > 0 {
		nearest := candidatesHeap.items[0]

		// Stop once the results are full and the nearest unexplored candidate is farther than the
		// current worst result: nothing closer can be reached from here.
		if results.Len() >= ef && nearest.dist > results.items[0].dist {
			break
		}
		heap.Pop(candidatesHeap)

		node, ok, err := g.store.GetNode(nearest.id)
		if err != nil {
			return nil, err
		}
		if !ok {
			// The frontier can hold an id a reclaim pass removed since it was queued. Skip it.
			continue
		}

		var neighbours []NodeID
		if layer < len(node.Layers) {
			neighbours = node.Layers[layer]
		}

		for _, nb := range neighbours {
			if _, seen := visited[nb]; seen {
				continue
			}
			visited[nb] = struct{}{}

			nbNode, ok, err := g.store.GetNode(nb)
			if err != nil {
				return nil, err
			}
			if !ok {
				// A neighbour list can name a node a reclaim pass removed. Skip it.
				continue
			}

			d := distance(g.metric, query, nbNode.Vector)

			if results.Len() < ef || d < results.items[0].dist {
				heap.Push(candidatesHeap, candidate{id: nb, dist: d, vector: nbNode.Vector})
				if !nbNode.Deleted {
					heap.Push(results, candidate{id: nb, dist: d, vector: nbNode.Vector})
					if results.Len() > ef {
						heap.Pop(results)
					}
				}
			}
		}
	}

	// results is a heap, so only items[0] is ordered; the rest of the slice is not. Sort it to return
	// the candidates nearest-first.
	out := make([]candidate, len(results.items))
	copy(out, results.items)
	sort.Slice(out, func(i, j int) bool { return out[i].dist < out[j].dist })
	return out, nil
}

// selectNeighborsHeuristic implements SELECT-NEIGHBORS-HEURISTIC
// (Algorithm 4): given a candidate set (each carrying its own vector) and
// a target count m, it greedily selects candidates nearest-first to the
// query, accepting a candidate e only if it is closer to the query than
// it is to every already-selected neighbour (the RNG/Delaunay
// diversification condition: dist(e, q) < dist(e, r) for all selected r).
// This produces a more diverse neighbour set than naive closest-M
// selection, which materially improves recall.
//
// Note: the paper's heuristic also defines optional extendCandidates and
// keepPrunedConnections behaviours; they are not implemented here for
// simplicity (MVP), but could be added later if desired.
func selectNeighborsHeuristic(metric Metric, candidates []candidate, m int) []candidate {
	sorted := make([]candidate, len(candidates))
	copy(sorted, candidates)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].dist < sorted[j].dist })

	selected := make([]candidate, 0, m)
	for _, e := range sorted {
		if len(selected) >= m {
			break
		}
		diverse := true
		for _, r := range selected {
			if distance(metric, e.vector, r.vector) < e.dist {
				diverse = false
				break
			}
		}
		if diverse {
			selected = append(selected, e)
		}
	}
	return selected
}
