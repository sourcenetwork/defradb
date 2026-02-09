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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// This test asserts that relational documents do not fail to sync if their related
// document does not exist at the destination.
func TestP2POneToManyPeerWithCreateUpdateLinkingSyncedDocToUnsyncedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
				// Create Saadi on first node
				// NodePeers do not sync new documents so this will not be synced
				// to node 1.
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"Name": "Saadi"
				}`,
			},
			testUtils.CreateDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(1, 0),
				},
			},
			testUtils.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				Doc: `{
					"_AuthorID": "bae-9ace7ed9-8229-5d2f-9e30-ffd5d2c84406"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Book {
						Name
						_AuthorID
						Author {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"Name":      "Gulistan",
							"_AuthorID": testUtils.NewDocIndex(0, 0),
							"Author": map[string]any{
								"Name": "Saadi",
							},
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Book {
						Name
						_AuthorID
						Author {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"Name":      "Gulistan",
							"_AuthorID": testUtils.NewDocIndex(0, 0),
							// "Saadi" was not synced to node 1, the update did not
							// result in an error and synced to relational id even though "Saadi"
							// does not exist in this node.
							"Author": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
