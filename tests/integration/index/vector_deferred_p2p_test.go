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

// Documents written, updated and deleted on one peer reach the other peer's deferred vector index
// through the merge path, which calls the same index methods as a local write. Results are the same
// whether the replica's entries are still pending or already applied to the graph.
func TestVectorIndexDeferredP2P_ReplicatedWrites_AreSearchableOnReplica(t *testing.T) {
	top2 := func(results ...map[string]any) *action.Request {
		return &action.Request{
			NodeID: immutable.Some(1),
			Request: `query {
				Users(order: {_alias: {sim: DESC}}, limit: 2){
					name
					sim: SIMILARITY(vector: {vector: [1, 0, 0]})
				}
			}`,
			Results: map[string]any{"Users": results},
		}
	}

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
					vector: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
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
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{"name": "near", "vector": []float32{1, 0, 0}},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{"name": "mid", "vector": []float32{0.9, 0.4, 0}},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{"name": "far", "vector": []float32{0, 1, 0}},
			},
			testUtils.WaitForSync{},
			top2(
				map[string]any{"name": "near", "sim": cosX(1, 0, 0)},
				map[string]any{"name": "mid", "sim": cosX(0.9, 0.4, 0)},
			),
			// "far" moves next to the query on node 0 and the replica sees the update.
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  2,
				Doc:    `{"vector": [0.99, 0.01, 0]}`,
			},
			testUtils.WaitForSync{},
			top2(
				map[string]any{"name": "near", "sim": cosX(1, 0, 0)},
				map[string]any{"name": "far", "sim": cosX(0.99, 0.01, 0)},
			),
			// Deleting the nearest on node 0 removes it from the replica's results.
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			top2(
				map[string]any{"name": "far", "sim": cosX(0.99, 0.01, 0)},
				map[string]any{"name": "mid", "sim": cosX(0.9, 0.4, 0)},
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
