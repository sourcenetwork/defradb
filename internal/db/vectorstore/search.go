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

// SearchResult is one vector-search hit. A smaller Distance is nearer.
type SearchResult struct {
	DocID    string
	Distance float64
}

// Search returns up to k documents nearest to query, nearest first, reading through the transaction
// on ctx.
//
// A hit whose short id no longer maps to a document is skipped: the document can be deleted in this
// same transaction after the graph was read, and a dangling id must not reach the caller.
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
