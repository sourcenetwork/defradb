// Copyright 2023 Democratized Data Foundation
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
func TestP2PWithMultipleDocumentsSingleDelete(t *testing.T) {
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
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithMultipleDocumentsSingleDeleteWithShowDeleted(t *testing.T) {
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
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
					{
						"_deleted": true,
						"Name":     "John",
						"Age":      int64(43),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithMultipleDocumentsWithSingleUpdateBeforeConnectSingleDeleteWithShowDeleted(t *testing.T) {
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
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 60
				}`,
				DontSync: true,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
					{
						"_deleted": true,
						"Name":     "John",
						"Age":      int64(60),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithMultipleDocumentsWithMultipleUpdatesBeforeConnectSingleDeleteWithShowDeleted(t *testing.T) {
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
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 60
				}`,
				DontSync: true,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 62
				}`,
				DontSync: true,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
					{
						"_deleted": true,
						"Name":     "John",
						"Age":      int64(62),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PWithMultipleDocumentsWithUpdateAndDeleteBeforeConnectSingleDeleteWithShowDeleted(t *testing.T) {
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
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 60
				}`,
				DontSync: true,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 62
				}`,
				DontSync: true,
			},
			testUtils.DeleteDoc{
				NodeID:   immutable.Some(0),
				DocID:    0,
				DontSync: true,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				// Update John's Age on the second node only
				NodeID: immutable.Some(1),
				DocID:  0,
				Doc: `{
					"Age": 66
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
					{
						"_deleted": true,
						"Name":     "John",
						"Age":      int64(62),
					},
				},
			},
			// The target node currently won't receive the pre-connection updates from the source.
			// We should look into adding a head exchange mechanic on connect.
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      int64(74),
					},
					{
						"_deleted": false,
						"Name":     "John",
						"Age":      int64(66),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
