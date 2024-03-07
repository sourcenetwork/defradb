// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexP2P_IfPeerCreatedDoc_ListeningPeerShouldIndexIt(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Fred"}}){
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Fred",
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
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Islam"}}){
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Islam",
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
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.CreateDoc{
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
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users (filter: {name: {_eq: "Fred"}}){
						age
					}
				}`,
				Results: []map[string]any{
					{
						"age": int64(30),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
