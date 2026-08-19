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
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
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
	// readyVectorIndexOnField only returns a vector index, so GetVector always succeeds here.
	vectorDesc, _ := index.GetVector()

	query, ok := similarityQueryVector(sim.Vector)
	if !ok {
		return nil
	}

	// A wrong-length query would be scored on only its shared leading elements, giving wrong results.
	// The full-scan path errors on this; do the same here.
	if dims := int(vectorDesc.Dimensions); dims > 0 && len(query) != dims {
		return NewErrMismatchLengthOnSimilarity(dims, len(query))
	}

	// Offset skips the first documents of the result, so the graph must return them too, not just the
	// Limit that remains after skipping. Ask for Limit+Offset; the limit node applies the offset.
	k := int(n.selectReq.Limit.Limit) + int(n.selectReq.Limit.Offset)
	prefixes, err := n.vectorSearchPrefixes(index, query, k)
	if err != nil {
		return err
	}

	origScan.Prefixes(prefixes)
	origScan.vectorIndexed = true
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
//
// Any metric qualifies, because similarityNode scores by the index's metric, so the k documents the
// graph returns are the k the query asked for.
func (n *selectNode) readyVectorIndexOnField(fieldName string) (client.IndexDescription, bool) {
	for _, idx := range queryableIndexesOnField(n.collection, fieldName) {
		if idx.IsVector() {
			return idx, true
		}
	}
	return client.IndexDescription{}, false
}

// vectorIndexMetricOnField returns the metric of the field's vector index, or cosine if it has no
// index. Cosine is the default because that is what `_similarity` meant before any metric existed, so
// an unindexed field keeps scoring as it did.
func vectorIndexMetricOnField(col client.Collection, fieldName string) client.DistanceMetric {
	for _, idx := range queryableIndexesOnField(col, fieldName) {
		if vector, ok := idx.GetVector(); ok {
			return vector.Metric
		}
	}
	return client.DistanceMetricCosine
}

// vectorSearchPrefixes runs the graph search and returns one datastore prefix per nearest document.
func (n *selectNode) vectorSearchPrefixes(
	index client.IndexDescription,
	query []float32,
	k int,
) ([]keys.Walkable, error) {
	// This is only reached for a vector index (the caller already checked), so GetVector always succeeds.
	vectorDesc, _ := index.GetVector()

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

	results, err := vectorindex.Search(
		n.planner.ctx, collectionShortID, index.ID, epoch, *vectorDesc, query, k,
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
