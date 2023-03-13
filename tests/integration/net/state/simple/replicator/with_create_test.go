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

// TestP2FullPReplicator tests document syncing between a node and a replicator.
func TestP2POneToOneReplicator(t *testing.T) {
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
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
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
						"Age": uint64(21),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2POneToOneReplicatorDoesNotSyncExisting(t *testing.T) {
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
				// Create John on the first (source) node only
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			// Once configured the replicator should sync existing documents
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
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
						"Age": uint64(21),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2POneToManyReplicator(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
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
						"Age": uint64(21),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2POneToOneOfManyReplicator(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Node[2] will not be configured
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
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Age": uint64(21),
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Age": uint64(21),
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(2),
				Request: `query {
					Users {
						Age
					}
				}`,
				// As node[2] was not configured, John should not be synced to it
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2POneToOneReplicatorManyDocs(t *testing.T) {
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
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				// Create Fred on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 22
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
						"Age": uint64(21),
					},
					{
						"Age": uint64(22),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2POneToManyReplicatorManyDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				// Create Fred on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 22
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
						"Age": uint64(21),
					},
					{
						"Age": uint64(22),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
