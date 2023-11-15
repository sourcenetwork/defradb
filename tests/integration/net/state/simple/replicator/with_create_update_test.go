// Copyright 2022 Democratized Data Foundation
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

func TestP2POneToOneReplicatorWithCreateWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// This document is created in node `0` after the replicator has
				// been set up. Its creation and future updates should be synced
				// across all configured nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Age": int64(60),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorWithCreateWithUpdateOnRecipientNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// This document is created in node `0` after the replicator has
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
			testUtils.UpdateDoc{
				// Update John's Age on the seond node only, and allow the value to sync
				// back to the original node that created the document.
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Age": int64(60),
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
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				// This document is created in all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConfigureReplicator{
				// Replication must happen after creating documents
				// on both nodes, or a race condition can occur
				// on the second node when creating the document
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// This document is created in the second node (target) only
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "Fred",
					"Age": 40
				}`,
			},
			testUtils.UpdateDoc{
				// Update Fred's Age
				NodeID: immutable.Some(1),
				DocID:  1,
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Assert that the target node only contains John
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
