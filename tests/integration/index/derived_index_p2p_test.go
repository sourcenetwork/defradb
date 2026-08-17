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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// Replication runs through the same collection index maintenance hooks as local writes. Exercise
// add, update, and delete against both derived layouts so replicas never retain stale candidates or
// scores.
func TestDerivedIndexesP2P_ReplicatedLifecycleMaintainsTrigramAndBM25(t *testing.T) {
	replica := immutable.Some(1)
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{SDL: `
				type Users {
					name: String @trigramIndex
					bio: String @fullTextIndex
				}
			`},
			testUtils.ConnectPeers{SourceNodeID: 1, TargetNodeID: 0},
			testUtils.AddCollectionSubscription{NodeID: 1, CollectionIDs: []int{0}},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": "Alice", "bio": "database indexing"}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: replica,
				Request: `query {
					byName: Users(filter: {name: {_like: "Ali%"}}) { name }
					ranked: Users { name  rank: _bm25(query: "indexing", fields: ["bio"]) }
				}`,
				Results: map[string]any{
					"byName": []map[string]any{{"name": "Alice"}},
					"ranked": []map[string]any{{"name": "Alice", "rank": testUtils.BeNumerically(">", 0)}},
				},
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc:    `{"name": "Bob", "bio": "replication gossip"}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: replica,
				Request: `query {
					oldName: Users(filter: {name: {_like: "Ali%"}}) { name }
					byName: Users(filter: {name: {_regex: "^Bob"}}) { name }
					oldRank: Users { rank: _bm25(query: "indexing", fields: ["bio"]) }
					ranked: Users { name  rank: _bm25(query: "replication", fields: ["bio"]) }
				}`,
				Results: map[string]any{
					"oldName": []map[string]any{},
					"byName":  []map[string]any{{"name": "Bob"}},
					"oldRank": []map[string]any{},
					"ranked":  []map[string]any{{"name": "Bob", "rank": testUtils.BeNumerically(">", 0)}},
				},
			},
			testUtils.DeleteDoc{NodeID: immutable.Some(0), DocID: 0},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: replica,
				Request: `query {
					byName: Users(filter: {name: {_regex: "^Bob"}}) { name }
					ranked: Users { rank: _bm25(query: "replication", fields: ["bio"]) }
				}`,
				Results: map[string]any{
					"byName": []map[string]any{},
					"ranked": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
