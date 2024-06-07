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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestP2FullPReplicator tests document syncing between a node and a replicator.
func TestP2POneToManyReplicator(t *testing.T) {
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create Saadi on the first node
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"Name": "Saadi"
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				// Create Gulistan on the first node
				CollectionID: 1,
				Doc: `{
					"Name": "Gulistan",
					"Author_id": "bae-cf278a29-5680-565d-9c7f-4c46d3700cf0"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Both Saadi and Gulistan should be synced to all nodes and linked correctly
				Request: `query {
					Book {
						Name
						Author {
							Name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "Gulistan",
						"Author": map[string]any{
							"Name": "Saadi",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
