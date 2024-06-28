// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test asserts that relational documents do not fail to sync if their related
// document does not exist at the destination.
func TestP2POneToManyPeerWithCreateUpdateLinkingSyncedDocToUnsyncedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						Name: String
						Books: [Book]
					}
					type Book {
						Name: String
						Author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				// Create Gulistan on all nodes
				CollectionID: 1,
				Doc: `{
					"Name": "Gulistan"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create Saadi on first node
				// NodePeers do not sync new documents so this will not be synced
				// to node 1.
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"Name": "Saadi"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				Doc: `{
					"Author_id": "bae-6a4c24c0-7b0b-5f51-a274-132d7ca90499"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Book {
						Name
						Author_id
						Author {
							Name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"Name":      "Gulistan",
						"Author_id": testUtils.NewDocIndex(0, 0),
						"Author": map[string]any{
							"Name": "Saadi",
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Book {
						Name
						Author_id
						Author {
							Name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"Name":      "Gulistan",
						"Author_id": testUtils.NewDocIndex(0, 0),
						// "Saadi" was not synced to node 1, the update did not
						// result in an error and synced to relational id even though "Saadi"
						// does not exist in this node.
						"Author": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
