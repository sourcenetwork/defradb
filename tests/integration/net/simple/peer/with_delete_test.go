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

package peer_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestP2PWithSingleDocumentConcurrentDeleteAndUpdate(t *testing.T) {
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
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Jane"
				}`,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(1),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_deleted": true,
							"Name":     testUtils.AnyOf("John", "Jane"),
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
func TestP2PWithMultipleDocumentsSingleDelete(t *testing.T) {
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
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			&action.AddDoc{
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
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_deleted": false,
							"Name":     "Andy",
							"Age":      int64(74),
						},
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
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			&action.AddDoc{
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
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
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
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			&action.AddDoc{
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
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
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
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			&action.AddDoc{
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
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 62
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
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
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			&action.AddDoc{
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
			},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"Age": 62
				}`,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 1),
				},
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
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
			},
			// The target node currently won't receive the pre-connection updates from the source.
			// We should look into adding a head exchange mechanic on connect.
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
