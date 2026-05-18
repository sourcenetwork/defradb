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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestP2POneToOneReplicatorWithAddWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// This document is added in node `0` after the replicator has
				// been set up. Its creation and future updates should be synced
				// across all configured nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.UpdateDoc{
				// Update John's Age on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorWithAddWithUpdateOnRecipientNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// This document is added in node `0` after the replicator has
				// been set up. Its creation and future updates should be synced
				// across all configured nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			// Wait for John to be synced to the target before attempting to update
			// it.
			testUtils.WaitForSync{},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			&action.UpdateDoc{
				// Update John's Age on the seond node only, and allow the value to sync
				// back to the original node that added the document.
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorDoesNotUpdateDocExistingOnlyOnTarget(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// This document is added in all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.AddReplicator{
				// Replication must happen after adding documents
				// on both nodes, or a race condition can occur
				// on the second node when adding the document
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				// This document is added in the second node (target) only
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "Fred",
					"Age": 40
				}`,
			},
			&action.UpdateDoc{
				// Update Fred's Age
				NodeID: immutable.Some(1),
				DocID:  1,
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// Assert that the target node only contains John
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
