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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				// Create Saadi on the first node
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `{
					"Name": "Saadi"
				}`,
			},
			&action.AddDoc{
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
