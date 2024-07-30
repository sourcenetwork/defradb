// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// The parent-child distinction in these tests is as much documentation and test
// of the test system as of production.  See it as a santity check of sorts.
func TestP2PWithSingleDocumentSingleUpdateFromChild(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
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

// The parent-child distinction in these tests is as much documentation and test
// of the test system as of production.  See it as a santity check of sorts.
func TestP2PWithSingleDocumentSingleUpdateFromParent(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the second node only, and allow the value to sync
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

// TestP2PWithSingleDocumentUpdatePerNode tests document syncing between two nodes with a single update per node
func TestP2PWithSingleDocumentUpdatePerNode(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node to 60
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the second node to 45
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": testUtils.AnyOf{int64(45), int64(60)},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithSingleDocumentSingleUpdateDoesNotSyncToNonPeerNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// This last node is not marked for peer sync
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node to 60
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(60),
						},
					},
				},
			},
			testUtils.Request{
				// Update should not be synced to this node
				NodeID: immutable.Some(2),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithSingleDocumentSingleUpdateDoesNotSyncFromUnmappedNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// This node is unmapped, updates applied to this node should
			// not be synced to the other nodes.
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the third (unmapped) node to 60
				NodeID: immutable.Some(2),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Update should not be synced to this node
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
			testUtils.Request{
				// Update should not be synced to this node
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
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

// TestP2PWithMultipleDocumentUpdatesPerNode tests document syncing between two nodes with multiple updates per node.
func TestP2PWithMultipleDocumentUpdatesPerNode(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 61
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 62
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 46
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 47
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": testUtils.AnyOf{int64(47), int64(62)},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestP2PWithSingleDocumentSingleUpdateFromChildWithP2PCollection tests document syncing between two nodes by
// ensuring that the created document and its update reach the node that subscribes to the P2P collection topic.
func TestP2PWithSingleDocumentSingleUpdateFromChildWithP2PCollection(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 31
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  1,
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
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

// TestP2PWithMultipleDocumentUpdatesPerNodeWithP2PCollection tests document syncing between two nodes with multiple
// updates per node as well as ensuring that the created document and its update reach the node that
// subscribes to the P2P collection topic.
func TestP2PWithMultipleDocumentUpdatesPerNodeWithP2PCollection(t *testing.T) {
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
				// Create Shahzad on all nodes
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred",
					"Age": 31
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 60
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 61
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 61
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 62
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 62
				}`,
			},
			testUtils.UpdateDoc{
				// Update Fred to 60
				NodeID: immutable.Some(0),
				DocID:  1,
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 45 on second node
				NodeID: immutable.Some(1),
				DocID:  0,
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 46 on second node
				NodeID: immutable.Some(1),
				DocID:  0,
				Doc: `{
					"Age": 46
				}`,
			},
			testUtils.UpdateDoc{
				// Update John to 47 on second node
				NodeID: immutable.Some(1),
				DocID:  0,
				Doc: `{
					"Age": 47
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": testUtils.AnyOf{int64(47), int64(62)},
						},
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
