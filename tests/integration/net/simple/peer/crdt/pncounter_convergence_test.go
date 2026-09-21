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
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestP2PUpdate_WithPNCounterRepeatedSimultaneousUpdates_Converges asserts that a pncounter
// converges to the sum of every applied increment when both peers are updated repeatedly.
//
// A pncounter is a state-based CRDT: each increment is a distinct delta block, and the merged
// value must equal the sum of all deltas regardless of the order or interleaving in which the
// blocks arrive at each node.
//
// Here each of the two nodes applies 20 increments of +1 to a counter starting at 0, so both
// nodes must converge on 40.
//
// Without the accompanying fix both nodes hold an identical set of delta blocks but materialise
// different values, both greater than 40, because an increment is applied more than once during
// merge. The wrong value is persisted, so it survives a node restart.
//
// This is not caused by dropped or undelivered blocks: it reproduces with an identical block set
// on both nodes and no transport errors logged. Nor is it a regression - it reproduces
// identically on 63fa24b85, the commit preceding "refactor: Detangle crdts" (#5163).
func TestP2PUpdate_WithPNCounterRepeatedSimultaneousUpdates_Converges(t *testing.T) {
	const updatesPerNode = 20
	const expectedPoints = int64(2 * updatesPerNode)

	actions := []any{
		testUtils.RandomNetworkingConfig(),
		testUtils.RandomNetworkingConfig(),
		&action.AddCollection{
			SDL: `
				type Users {
					name: String
					points: Int @crdt(type: pncounter)
				}
			`,
		},
		&action.AddDoc{
			// Create John on all nodes
			Doc: `{
				"name": "John",
				"points": 0
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
	}

	// Interleave the updates across both nodes so that each node is producing deltas whilst
	// also merging the deltas produced by its peer.
	for range updatesPerNode {
		actions = append(actions,
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"points": 1
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				DocID:  0,
				Doc: `{
					"points": 1
				}`,
			},
		)
	}

	actions = append(actions,
		testUtils.WaitForSync{},
		&action.Request{
			Request: `query {
				Users {
					points
				}
			}`,
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"points": expectedPoints,
					},
				},
			},
		},
	)

	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		//
		// Signing makes each node's genesis block (and thus DocID) signer-specific, so creating the
		// doc on every node yields distinct docs that never converge.
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.SignedDocs},
		Actions:            actions,
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestP2PUpdate_WithPNCounterFiveNodesRepeatedUpdates_AllConverge asserts that every node in a
// five node mesh ends up on the same counter value, and that the value is the sum of every
// increment applied anywhere in the mesh.
//
// Each of the five nodes applies its increments while merging those of the other four, which is
// what causes an already-counted composite to be collected again during a merge.
//
// The request at the end carries no NodeID, so it is asserted against all five nodes: it fails if
// any node disagrees with the others, and it fails if they agree on the wrong total.
//
// See TestP2PUpdate_WithPNCounterRepeatedSimultaneousUpdates_Converges for the two node case and
// the background on why the counter is applied more than once.
func TestP2PUpdate_WithPNCounterFiveNodesRepeatedUpdates_AllConverge(t *testing.T) {
	const nodeCount = 5
	const updatesPerNode = 10
	const expectedPoints = int64(nodeCount * updatesPerNode)

	actions := []any{}

	for range nodeCount {
		actions = append(actions, testUtils.RandomNetworkingConfig())
	}

	actions = append(actions,
		&action.AddCollection{
			SDL: `
				type Users {
					name: String
					points: Int @crdt(type: pncounter)
				}
			`,
		},
		&action.AddDoc{
			// Create John on all nodes
			Doc: `{
				"name": "John",
				"points": 0
			}`,
		},
	)

	// Connect every node to every other node, so an update made on any node can reach all of the
	// others without having to be relayed.
	for source := range nodeCount {
		for target := source + 1; target < nodeCount; target++ {
			actions = append(actions, testUtils.ConnectPeers{
				SourceNodeID: source,
				TargetNodeID: target,
			})
		}
	}

	for nodeID := range nodeCount {
		actions = append(actions, testUtils.AddDocumentSubscription{
			NodeID: nodeID,
			DocIDs: []state.ColDocIndex{
				state.NewColDocIndex(0, 0),
			},
		})
	}

	// Interleave the updates so that every node is producing increments whilst merging those of
	// the other four.
	for range updatesPerNode {
		for nodeID := range nodeCount {
			actions = append(actions, &action.UpdateDoc{
				NodeID: immutable.Some(nodeID),
				DocID:  0,
				Doc: `{
					"points": 1
				}`,
			})
		}
	}

	actions = append(actions,
		testUtils.WaitForSync{},
		&action.Request{
			// No NodeID, so this is asserted against every node in the mesh.
			Request: `query {
				Users {
					points
				}
			}`,
			Results: map[string]any{
				"Users": []map[string]any{
					{
						"points": expectedPoints,
					},
				},
			},
		},
	)

	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		//
		// Signing makes each node's genesis block (and thus DocID) signer-specific, so creating the
		// doc on every node yields distinct docs that never converge.
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.SignedDocs},
		Actions:            actions,
	}

	testUtils.ExecuteTestCase(t, test)
}
