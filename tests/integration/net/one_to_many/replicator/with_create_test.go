// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replicator

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestP2POneToManyReplicator tests document syncing between a node and a replicator.
func TestP2POneToManyReplicator(t *testing.T) {
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
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.CreateDoc{
				// Create Saadi on the first node
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"Name": "Saadi"
				}`,
			},
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				// Create Gulistan on the first node
				CollectionID: 1,
				Doc: `{
					"Name": "Gulistan",
					"_AuthorID": "bae-9ace7ed9-8229-5d2f-9e30-ffd5d2c84406"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Both Saadi and Gulistan should be synced to all nodes and linked correctly
				Request: `query {
					Book {
						Name
						Author {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"Name": "Gulistan",
							"Author": map[string]any{
								"Name": "Saadi",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
