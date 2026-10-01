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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestIndexP2P_IfPeerAddedDoc_ListeningPeerShouldIndexIt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
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
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Fred"}}){
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexP2P_IfPeerUpdateDoc_ListeningPeerShouldUpdateIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
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
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Islam"}}){
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Islam",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexP2P_IfPeerDeleteDoc_ListeningPeerShouldDeleteIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
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
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred",
					"age": 30
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Fred"}}){
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age": int64(30),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The first push of "Bob" is rejected, so it ends up in the retry queue together with "Alice",
// which is always rejected. "Bob" can only arrive if "Alice" does not block the retry.
func TestIndexP2P_UniqueConflictOnReplicatorRetry_ShouldNotBlockOtherDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index(unique: true)
						age: Int
					}
				`,
			},
			&action.AddDoc{
				// Only exists on the target, its twin on the source can never be merged.
				NodeID: immutable.Some(1),
				Doc: `{
					"name": "Alice",
					"age": 1
				}`,
			},
			&action.AddDoc{
				// Only exists on the target, will be renamed to unblock its twin on the source.
				NodeID: immutable.Some(1),
				Doc: `{
					"name": "Bob",
					"age": 9
				}`,
			},
			&action.AddDoc{
				// Permanently conflicts with the doc on the target.
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Alice",
					"age": 2
				}`,
			},
			&action.AddDoc{
				// Conflicts with the doc on the target until it is renamed.
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 3
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				DocID:  1,
				Doc:    `{"name": "Xavier"}`,
			},
			testUtils.WaitForSync{
				// node 0's "Alice" is always rejected by node 1.
				ExcludedDocs: []state.ColDocIndex{state.NewColDocIndex(0, 2)},
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users(order: {name: ASC}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice", "age": int64(1)},
						{"name": "Bob", "age": int64(3)},
						{"name": "Xavier", "age": int64(9)},
					},
				},
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  3,
				Doc:    `{"age": 5}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users(order: {name: ASC}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice", "age": int64(1)},
						{"name": "Bob", "age": int64(5)},
						{"name": "Xavier", "age": int64(9)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
