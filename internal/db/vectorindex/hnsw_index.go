// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// hnswIndex is the HNSW implementation of Index: a graph for one vector index, bound to the
// transaction on ctx. It is the only type here that touches the engine, so if the engine is ever
// extracted this and hnsw_store.go are the whole seam.
type hnswIndex struct {
	graph    *hnsw.HNSWIndex
	efSearch int
}

var _ Index = (*hnswIndex)(nil)

// openHNSW builds the HNSW index described by desc over the given (collection, index, epoch)
// keyspace. It fails if desc asks for a metric the engine does not support, so an unusable index is
// rejected when it is opened rather than on the first write or search.
//
// The seed is the index id, not a random value, so layer choices are reproducible across runs, and
// so the write path and the read path build the same graph. The index id is stable and unique within
// its collection, so it serves as one.
func openHNSW(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	desc client.VectorIndexDescription,
) (*hnswIndex, error) {
	metric, err := hnswMetric(desc.Metric)
	if err != nil {
		return nil, err
	}

	// DefaultParams sets sane defaults (and guards M). Only override them when the descriptor gives a
	// positive value, so a descriptor missing EfConstruction/EfSearch stays usable instead of getting
	// a zero candidate list.
	params := hnsw.DefaultParams(int(desc.HNSW.M))
	if desc.HNSW.EfConstruction > 0 {
		params.EfConstruction = int(desc.HNSW.EfConstruction)
	}
	if desc.HNSW.EfSearch > 0 {
		params.EfSearch = int(desc.HNSW.EfSearch)
	}

	store := newNodeStore(ctx, collectionShortID, indexID, epoch)
	return &hnswIndex{
		graph:    hnsw.New(store, metric, params, int64(indexID)),
		efSearch: params.EfSearch,
	}, nil
}

// Insert adds or replaces the vector stored for nodeID.
func (i *hnswIndex) Insert(nodeID uint64, vec []float32) error {
	return i.graph.Insert(hnsw.NodeID(nodeID), vec)
}

// Delete removes nodeID from the search results (the engine tombstones it).
func (i *hnswIndex) Delete(nodeID uint64) error {
	return i.graph.Delete(hnsw.NodeID(nodeID))
}

// Search returns up to k hits nearest to query, nearest first.
func (i *hnswIndex) Search(query []float32, k int) ([]Hit, error) {
	neighbors, err := i.graph.SearchWithDistance(query, k, i.efSearch)
	if err != nil {
		return nil, err
	}
	hits := make([]Hit, len(neighbors))
	for j, n := range neighbors {
		hits[j] = Hit{NodeID: uint64(n.ID), Distance: n.Distance}
	}
	return hits, nil
}

// hnswMetric maps the descriptor's distance metric to the engine's. An unsupported metric is an
// error rather than a silent default, so a mis-configured index fails loudly.
func hnswMetric(m client.DistanceMetric) (hnsw.Metric, error) {
	switch m {
	case client.DistanceMetricCosine:
		return hnsw.Cosine, nil
	case client.DistanceMetricEuclidean:
		return hnsw.Euclidean, nil
	case client.DistanceMetricDotProduct:
		return hnsw.DotProduct, nil
	default:
		return 0, newErrUnsupportedMetric(m)
	}
}
