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

// TestP2PUpdate_WithPNCounterRepeatedSimultaneousUpdates_Converges asserts that a counter
// updated on both peers at once reaches the sum of every increment.
//
// Two nodes apply 20 increments of +1 each, so both must end on 40.
//
// Without the fix each node reads a value above 40, and the two disagree, even though they hold
// the same blocks. The wrong value is stored, so restarting a node does not clear it. Dropped
// blocks are not the cause: the block sets match and no transport errors are logged. It is also
// not a regression, reproducing the same way on 63fa24b85, before "refactor: Detangle crdts"
// (#5163).
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

	// Interleave the updates so each node is producing increments while merging its peer's.
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

// TestP2PUpdate_WithPNCounterFiveNodesRepeatedUpdates_AllConverge is the two node case widened to
// a five node mesh, where each node merges the increments of the other four.
//
// The closing request carries no NodeID, so it runs against all five nodes. It fails if any node
// disagrees with the others, and if they agree on the wrong total.
//
// See TestP2PUpdate_WithPNCounterRepeatedSimultaneousUpdates_Converges for why the counter is
// applied more than once.
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

	// Connect every pair, so an update reaches all the other nodes without being relayed.
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

	// Interleave the updates so every node is producing increments while merging the other four.
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
