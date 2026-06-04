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

package sync_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocSync_WithDocsAvailableOnSingleNode_ShouldSync(t *testing.T) {
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
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Andy",
					"Age": 25
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SyncDocs{
				NodeID:       1,
				CollectionID: 0,
				DocIDs:       []int{0, 1},
				SourceNodes:  []int{0, 0}, // Both documents are from node 0
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Andy",
							"Age":  int64(25),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSync_WithDocsAvailableOnMultipleNode_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
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
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "Andy",
					"Age": 25
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			testUtils.SyncDocs{
				NodeID:      2,
				DocIDs:      []int{0, 1},
				SourceNodes: []int{0, 1}, // Document 0 is from node 0, document 1 is from node 1
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(2),
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Andy",
							"Age":  int64(25),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSync_WithSingleDocAvailableOnMultipleNode_ShouldSync(t *testing.T) {
	addDocOnNode := func(nodeId int) *action.AddDoc {
		return &action.AddDoc{
			NodeID: immutable.Some(nodeId),
			Doc: `{
				"Name": "John",
				"Age": 21
			}`,
		}
	}

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
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
			addDocOnNode(0),
			addDocOnNode(1),
			addDocOnNode(2),
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 3,
			},
			testUtils.SyncDocs{
				NodeID:      3,
				DocIDs:      []int{0},
				SourceNodes: []int{0},
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(3),
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

const docSyncTopic = "doc-sync"

func TestDocSync_WithDifferentVersionsOnPeers_ShouldSyncLatest(t *testing.T) {
	test := testUtils.TestCase{
		FlakeRetries: 5,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
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
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 22
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(2),
				Doc: `{
					"Age": 23
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(3),
				Doc: `{
					"Age": 24
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 25
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(3),
				Doc: `{
					"Age": 26
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(3),
				Doc: `{
					"Age": 27
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 3,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					docSyncTopic: {1, 2, 3},
				},
			},
			testUtils.SyncDocs{
				NodeID:       0,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{1, 2, 3},
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSync_AfterSync_ShouldNotSubscribeToDocUpdates(t *testing.T) {
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
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SyncDocs{
				NodeID:       1,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{0},
			},
			testUtils.WaitForSync{},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 22
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
