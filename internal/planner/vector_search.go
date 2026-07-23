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

// tryRouteSimilarityToVectorIndex accelerates a nearest-neighbour query when there is a ready vector
// index for it. The shape it looks for is a `_similarity` on a vector field, ordered by that
// similarity descending, with a limit: this is "give me the k documents nearest to a query vector".
//
// When it matches, it runs the graph search for the k nearest documents and narrows the scan to just
// those documents (via prefixes, the same way a query by document id does). The rest of the plan is
// unchanged: the similarity node still scores each fetched document and the order/limit nodes still
// sort and cap them. So instead of scanning the whole collection, the scan reads only k documents.
//
// It does nothing (leaving the full-scan fallback in place) when the query does not have this exact
// shape, when there is no ready vector index on the field, or when there is a filter on the
// similarity (a filter could drop some of the k nearest and need more candidates than the graph was
// asked for; that is filtered KNN, tracked in issue #5071).
func (n *selectNode) tryRouteSimilarityToVectorIndex(origScan *scanNode) error {
	// A limit is what makes this a k-nearest query; without it there is no k to search for.
	if n.selectReq.Limit == nil || n.selectReq.Limit.Limit <= 0 {
		return nil
	}
	// A filter on the query means results may be dropped after the search, so the k nearest is not
	// necessarily enough. Leave it to the full-scan path (see #5071).
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

	targetField := sim.SimilarityTarget.Field.Name
	index, ok := n.readyVectorIndexOnField(targetField)
	if !ok {
		return nil
	}

	query, ok := similarityQueryVector(sim.Vector)
	if !ok {
		return nil
	}

	k := int(n.selectReq.Limit.Limit)
	prefixes, err := n.vectorSearchPrefixes(index, query, k)
	if err != nil {
		return err
	}

	origScan.Prefixes(prefixes)
	// No document was near enough to return (an empty graph, or every hit resolved to a deleted
	// document). Mark the scan empty so it does not fall back to a full prefix scan.
	if len(prefixes) == 0 {
		origScan.noResults = true
	}
	return nil
}

// singleSimilarityField returns the sole `_similarity` field in the selection, or nil if there is
// not exactly one. Routing a query with two similarities is ambiguous (which one is the search?), so
// those keep the full-scan path.
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

// isOrderedBySimilarityDesc reports whether the query orders by the given similarity field,
// descending, and by nothing else. Descending because a larger cosine similarity means nearer, and
// nothing else because a secondary sort key would need documents the graph search did not return.
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

// readyVectorIndexOnField returns the ready vector index on the named field, if there is one.
// queryableIndexesOnField already excludes indexes that are still building or have failed.
func (n *selectNode) readyVectorIndexOnField(fieldName string) (client.IndexDescription, bool) {
	for _, idx := range queryableIndexesOnField(n.collection, fieldName) {
		if idx.Kind() == client.IndexKindVector {
			return idx, true
		}
	}
	return client.IndexDescription{}, false
}

// vectorSearchPrefixes runs the graph search for the k nearest documents to query and returns them
// as datastore prefixes, one per document, so the scan reads only those documents.
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

// similarityQueryVector converts the query vector from the request (a slice of numbers typed as any)
// into a []float32 for the graph search. It returns ok == false if the value is not a numeric slice.
func similarityQueryVector(vector any) ([]float32, bool) {
	vec := convertArray[float32](vector)
	if vec == nil {
		return nil, false
	}
	return vec, true
}

// vectorIndexMetricAndParams reads the distance metric and HNSW parameters from the index
// description, the same way the write path builds them, so read and write search the same graph.
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
