// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorstore

import (
	"context"

	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// SearchResult is one hit from a vector search: the matched document and its distance to the query
// vector (smaller is nearer). The document id is the normal string id, resolved from the graph's
// internal node id.
type SearchResult struct {
	DocID    string
	Distance float64
}

// Search runs a k-nearest-neighbour search against the vector index's graph and returns up to k
// documents ordered nearest first. It reads through the transaction on ctx.
//
// The graph stores documents by their short id; each hit is resolved back to its string document id
// here. A hit whose short id no longer maps to a document (e.g. the document was deleted in this same
// transaction after the graph was read) is skipped rather than returned as a dangling id.
func Search(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	metric hnsw.Metric,
	params hnsw.Params,
	query []float32,
	k int,
) ([]SearchResult, error) {
	graph := NewGraph(ctx, collectionShortID, indexID, epoch, metric, params)

	hits, err := graph.SearchWithDistance(query, k, params.EfSearch)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(hits))
	for _, hit := range hits {
		docID, found, err := id.GetDocID(ctx, uint64(hit.ID))
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		results = append(results, SearchResult{DocID: docID, Distance: hit.Distance})
	}
	return results, nil
}
