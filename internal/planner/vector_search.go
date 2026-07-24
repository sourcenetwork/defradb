// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorstore"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// tryRouteSimilarityToVectorIndex narrows an otherwise-full scan to the k nearest documents when the
// query is a nearest-neighbour search (a single `_similarity` ordered descending, with a limit) and
// the field has a ready vector index. It feeds the graph search results to the scan as document
// prefixes; the similarity/order/limit nodes are left to score, sort and cap as usual. When the query
// does not match, it leaves the full-scan path in place.
func (n *selectNode) tryRouteSimilarityToVectorIndex(origScan *scanNode) error {
	if n.selectReq.Limit == nil || n.selectReq.Limit.Limit <= 0 {
		return nil
	}
	// A filter can drop some of the k nearest, so the graph would need to return more than k to
	// backfill. That is filtered KNN: https://github.com/sourcenetwork/defradb/issues/5071
	if n.filter != nil {
		return nil
	}

	sim := n.singleSimilarityField()
	if sim == nil {
		return nil
	}
	if !n.isOrderedBySimilarityDesc(sim) {
		return nil
	}

	index, ok := n.readyVectorIndexOnField(sim.SimilarityTarget.Field.Name)
	if !ok {
		return nil
	}

	query, ok := similarityQueryVector(sim.Vector)
	if !ok {
		return nil
	}

	prefixes, err := n.vectorSearchPrefixes(index, query, int(n.selectReq.Limit.Limit))
	if err != nil {
		return err
	}

	origScan.Prefixes(prefixes)
	// Empty prefixes would otherwise let the scan fall back to reading the whole collection.
	if len(prefixes) == 0 {
		origScan.noResults = true
	}
	return nil
}

// singleSimilarityField returns the sole `_similarity` field, or nil if there are none or several:
// with two, which one drives the search is ambiguous, so such a query keeps the full-scan path.
func (n *selectNode) singleSimilarityField() *mapper.Similarity {
	var found *mapper.Similarity
	for _, field := range n.selectReq.Fields {
		if sim, ok := field.(*mapper.Similarity); ok {
			if found != nil {
				return nil
			}
			found = sim
		}
	}
	return found
}

// isOrderedBySimilarityDesc requires ordering by this similarity alone, descending: descending
// because larger cosine means nearer, alone because a second sort key would need documents beyond
// the k the graph returns.
func (n *selectNode) isOrderedBySimilarityDesc(sim *mapper.Similarity) bool {
	if n.selectReq.OrderBy == nil || len(n.selectReq.OrderBy.Conditions) != 1 {
		return false
	}
	cond := n.selectReq.OrderBy.Conditions[0]
	if cond.Direction != mapper.DESC {
		return false
	}
	return len(cond.FieldIndexes) == 1 && cond.FieldIndexes[0] == sim.Field.Index
}

// readyVectorIndexOnField returns the vector index on the field, if any. queryableIndexesOnField has
// already excluded indexes that are still building or have failed, so a returned index is usable.
func (n *selectNode) readyVectorIndexOnField(fieldName string) (client.IndexDescription, bool) {
	for _, idx := range queryableIndexesOnField(n.collection, fieldName) {
		if idx.Kind() == client.IndexKindVector {
			return idx, true
		}
	}
	return client.IndexDescription{}, false
}

// vectorSearchPrefixes runs the graph search and returns one datastore prefix per nearest document.
func (n *selectNode) vectorSearchPrefixes(
	index client.IndexDescription,
	query []float32,
	k int,
) ([]keys.Walkable, error) {
	collectionShortID, err := id.GetCollectionShortID(n.planner.ctx, n.collection.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	epoch, err := fetcher.ReadIndexEpoch(
		n.planner.ctx,
		datastore.CtxMustGetTxn(n.planner.ctx),
		n.collection.Version().CollectionID,
		index.ID,
	)
	if err != nil {
		return nil, err
	}

	metric, params, err := vectorIndexMetricAndParams(index)
	if err != nil {
		return nil, err
	}

	results, err := vectorstore.Search(
		n.planner.ctx, collectionShortID, index.ID, epoch, metric, params, query, k,
	)
	if err != nil {
		return nil, err
	}

	prefixes := make([]keys.Walkable, 0, len(results))
	for _, r := range results {
		docShortID, found, err := id.GetDocShortID(n.planner.ctx, collectionShortID, r.DocID)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		prefixes = append(prefixes, keys.DataStoreKey{
			CollectionShortID: collectionShortID,
			DocShortID:        docShortID,
		})
	}
	return prefixes, nil
}

func similarityQueryVector(vector any) ([]float32, bool) {
	vec := convertArray[float32](vector)
	if vec == nil {
		return nil, false
	}
	return vec, true
}

// vectorIndexMetricAndParams must build the metric and params the same way the write path does, so
// search traverses the same graph that maintenance wrote.
func vectorIndexMetricAndParams(index client.IndexDescription) (hnsw.Metric, hnsw.Params, error) {
	var metric hnsw.Metric
	switch index.Vector.Metric {
	case client.DistanceMetricCosine:
		metric = hnsw.Cosine
	default:
		return 0, hnsw.Params{}, ErrUnsupportedVectorMetric
	}

	params := hnsw.DefaultParams(int(index.Vector.HNSW.M))
	params.EfConstruction = int(index.Vector.HNSW.EfConstruction)
	params.EfSearch = int(index.Vector.HNSW.EfSearch)
	return metric, params, nil
}
