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
