// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// A document written on one peer and synced to another is added to the replica's graph, so a
// similarity query on the replica finds it. This proves the P2P merge maintains the vector index,
// not just direct writes. The @vectorIndex is in the schema so both peers build it the same way.
func TestVectorIndexP2P_ReplicatedDoc_IsSearchableOnReplica(t *testing.T) {
	test := testUtils.TestCase{
		// The vectorIndex directive does not exist in the older release.
		// https://github.com/sourcenetwork/defradb/issues/5121
		MultiplierExcludes: []string{
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `type Users {
					name: String
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			// Written on node 0. "near" is on the query direction [1,0,0], "far" is off-axis, so the
			// nearest is unambiguous.
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{"name": "near", "vector": []float32{1, 0, 0}},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{"name": "far", "vector": []float32{0, 1, 0}},
			},
			testUtils.WaitForSync{},
			// On the replica, the query returns the synced nearest document.
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users(order: {_alias: {sim: DESC}}, limit: 1){
						name
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "near", "sim": testUtils.CosineSimilarity([]float64{1, 0, 0}, []float64{1, 0, 0})},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
