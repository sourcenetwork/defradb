// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package branchable_collection

import (
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Make sure peers have time for libp2p data exchange setup.
// https://github.com/sourcenetwork/defradb/issues/4208
var waitConnection = testUtils.Wait{Duration: 100 * time.Millisecond}

func TestBranchableCollectionSync_WithMultipleDocsInComplexLinkedNetwork_ShouldSyncAll(t *testing.T) {
	// Network topology:
	// Node 0 ──── Node 1 ──── Node 2
	//    │
	//    └─────── Node 3 ──── Node 4
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User @branchable {
						name: String
						origin: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name":   "John",
					"origin": "node0",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name":   "Islam",
					"origin": "node1",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(2),
				DocMap: map[string]any{
					"name":   "Fred",
					"origin": "node2",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(3),
				DocMap: map[string]any{
					"name":   "Shahzad",
					"origin": "node3",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(4),
				DocMap: map[string]any{
					"name":   "Andy",
					"origin": "node4",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 4,
			},
			waitConnection,
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					User {
						name
						origin
					}
				}`,
				Results: map[string]any{
					"User": gomega.ConsistOf(
						map[string]any{
							"name":   "John",
							"origin": "node0",
						},
						map[string]any{
							"name":   "Islam",
							"origin": "node1",
						},
						map[string]any{
							"name":   "Fred",
							"origin": "node2",
						},
						map[string]any{
							"name":   "Shahzad",
							"origin": "node3",
						},
						map[string]any{
							"name":   "Andy",
							"origin": "node4",
						},
					),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_WithMultipleDocumentHeadsReceivedFromPeers_ShouldSyncAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User @branchable {
						name: String
						origin: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name":   "Islam",
					"origin": "node1",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(2),
				DocMap: map[string]any{
					"name":   "Fred",
					"origin": "node2",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			testUtils.WaitForSync{},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name":   "John",
					"origin": "node0",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					User {
						name
						origin
					}
				}`,
				Results: map[string]any{
					"User": gomega.ConsistOf(
						map[string]any{
							"name":   "John",
							"origin": "node0",
						},
						map[string]any{
							"name":   "Islam",
							"origin": "node1",
						},
						map[string]any{
							"name":   "Fred",
							"origin": "node2",
						},
					),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollectionSync_WithDocumentsFromPeers_ShouldHaveIdenticalDAG(t *testing.T) {
	sameCid1 := testUtils.NewSameValue()
	sameCid2 := testUtils.NewSameValue()
	sameCid3 := testUtils.NewSameValue()
	sameCid4 := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User @branchable {
						name: String
						origin: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name":   "John",
					"origin": "node0",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name":   "Islam",
					"origin": "node1",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(2),
				DocMap: map[string]any{
					"name":   "Fred",
					"origin": "node2",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(3),
				DocMap: map[string]any{
					"name":   "Andy",
					"origin": "node3",
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 3,
			},
			waitConnection,
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 3,
			},
			waitConnection,
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 2,
			},
			&action.SyncBranchableCollection{
				NodeID: 3,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					_commits(fieldName: null) {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": sameCid1,
						},
						{
							"cid": sameCid2,
						},
						{
							"cid": sameCid3,
						},
						{
							"cid": sameCid4,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
