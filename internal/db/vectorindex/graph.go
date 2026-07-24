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

	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// NewGraph exists so the write path (maintaining the graph) and the read path (searching it) build
// it identically; a mismatch would make search traverse a graph unlike the one that was written.
//
// The seed is the index id, not a random value, so layer choices are reproducible across runs. The
// index id is stable and unique within its collection, so it serves as one.
func NewGraph(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	metric hnsw.Metric,
	params hnsw.Params,
) *hnsw.Graph {
	store := NewNodeStore(ctx, collectionShortID, indexID, epoch)
	return hnsw.New(store, metric, params, int64(indexID))
}
