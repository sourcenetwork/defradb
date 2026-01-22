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

/*
todo - these tests are too flaky and block the merging of PRs during the working day (EST)
They should be added back in as part of https://github.com/sourcenetwork/defradb/issues/4308
when their flakiness has at least been reduced to a tolerable level.

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const syncBranchableTopic = "sync-branchable"

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
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 4,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {1, 2, 3, 4},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			testUtils.WaitForSync{},
			&action.Request{
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
			&action.Request{
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
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 3,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {1, 2, 3},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			&action.WaitForPeersEvents{
				NodeID: 1,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {0, 2, 3},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			&action.WaitForPeersEvents{
				NodeID: 2,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {0, 1, 3},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 2,
			},
			&action.WaitForPeersEvents{
				NodeID: 3,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {0, 1, 2},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 3,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_eq: null}}) {
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

func TestBranchableCollectionSync_WithDocumentsFromPeersAndNewHeadAfterSync_ShouldHaveIdenticalDAG(t *testing.T) {
	sameCid1 := testUtils.NewSameValue()
	sameCid2 := testUtils.NewSameValue()
	sameCid3 := testUtils.NewSameValue()
	sameCid4 := testUtils.NewSameValue()
	sameCid5 := testUtils.NewSameValue()

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
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 3,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {1},
				},
			},
			&action.WaitForPeersEvents{
				NodeID: 2,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {3},
				},
			},
			// We want to sync first node 0 with node 1 and node 2 with node 3 isolated to make sure
			// all nodes don't sync from the same source node
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			&action.SyncBranchableCollection{
				NodeID: 2,
			},
			testUtils.WaitForSync{},
			// Now connect all nodes together and sync.
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 3,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 3,
			},
			&action.WaitForPeersEvents{
				NodeID: 0,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {2, 3},
				},
			},
			&action.WaitForPeersEvents{
				NodeID: 1,
				ExpectedPeersByTopic: map[string][]int{
					syncBranchableTopic: {2, 3},
				},
			},
			&action.SyncBranchableCollection{
				NodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID: 3,
			},
			// Sync again on nodes 0 and 2 to get any missing heads (they have 2 heads each from previous syncs)
			&action.SyncBranchableCollection{
				NodeID: 0,
			},
			&action.SyncBranchableCollection{
				NodeID: 2,
			},
			testUtils.WaitForSync{},
			// Create another doc on all nodes to make sure it picked up all heads properly
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Bruno",
					"origin": "all",
				},
			},
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_eq: null}}) {
						cid
						height
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    sameCid1,
							"height": 2,
						},
						{
							"cid":    sameCid2,
							"height": 1,
						},
						{
							"cid":    sameCid3,
							"height": 1,
						},
						{
							"cid":    sameCid4,
							"height": 1,
						},
						{
							"cid":    sameCid5,
							"height": 1,
						},
					},
				},
			},
			// Make sure the new collection block for the new doc has all previous heads as links
			&action.Request{
				Request: `query {
					_commits(filter: {fieldName: {_eq: null}}, order: {height: DESC}, limit: 1) {
						heads {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"heads": gomega.HaveLen(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
*/
