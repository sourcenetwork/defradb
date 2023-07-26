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

	testUtils.ExecuteTEMP(t, test)
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

	testUtils.ExecuteTEMP(t, test)
}

func TestP2POneToOneReplicatorDoesNotSyncFromTargetToSource(t *testing.T) {
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
				// Create John on the second (target) node only
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Assert that John has not been synced to the first (source) node
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
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

	testUtils.ExecuteTEMP(t, test)
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

	testUtils.ExecuteTEMP(t, test)
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

	testUtils.ExecuteTEMP(t, test)
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

	testUtils.ExecuteTEMP(t, test)
}

func TestP2POneToOneReplicatorOrderIndependent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				NodeID: immutable.Some(0),
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.SchemaUpdate{
				// Add the same schema to the second node but with the age and name fields in
				// a different order.
				NodeID: immutable.Some(1),
				Schema: `
					type Users {
						age: Int
						name: String
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
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// The document should have been synced, and should contain the same values
				// including dockey and schema version id.
				Request: `query {
					Users {
						_key
						age
						name
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"age":  uint64(21),
						"name": "John",
						"_version": []map[string]any{
							{
								"schemaVersionId": "bafkreidovoxkxttybaew2qraoelormm63ilutzms7wlwmcr3xru44hfnta",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestP2POneToOneReplicatorOrderIndependentDirectCreate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				NodeID: immutable.Some(0),
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.SchemaUpdate{
				// Add the same schema to the second node but with the age and name fields in
				// a different order.
				NodeID: immutable.Some(1),
				Schema: `
					type Users {
						age: Int
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				// Create the document directly and indepentently on each node.
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				// Assert that the dockey and schema version id are the same across all nodes,
				// even though the schema field order is different.
				Request: `query {
					Users {
						_key
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"_version": []map[string]any{
							{
								"schemaVersionId": "bafkreidovoxkxttybaew2qraoelormm63ilutzms7wlwmcr3xru44hfnta",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
