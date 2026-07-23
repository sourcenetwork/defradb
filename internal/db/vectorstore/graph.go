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

	"github.com/sourcenetwork/defradb/internal/index/hnsw"
)

// NewGraph builds the HNSW graph for one vector index, reading and writing through the transaction
// on ctx. Both the write path (maintaining the graph) and the read path (searching it) call this so
// they always build the graph the same way.
//
// The seed is the index id rather than a random value, so inserts make the same random layer choices
// every run. The index id is stable and unique within its collection, so it works as a seed.
func NewGraph(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	metric hnsw.Metric,
	params hnsw.Params,
) *hnsw.Graph {
	store := NewNodeStore(ctx, collectionShortID, indexID, epoch)
	return hnsw.New(store, metric, params, int64(indexID))
}
