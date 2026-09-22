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

// The request carries no NodeID so it runs against both nodes, which matters because the
// failure this covers had them disagreeing, not just reading the wrong total.
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

	// Interleaved so each node is producing increments while merging its peer's. Sequential
	// updates never fork the chain and the bug does not appear.
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
		// Counter fields cannot be indexed:
		// https://github.com/sourcenetwork/defradb/issues/4439
		// Signing gives each node its own genesis block, so the DocIDs never match.
		// v1.0.0 still has this bug, so that node reports an inflated total.
		MultiplierExcludes: []string{
			multiplier.SecondaryIndex,
			multiplier.SignedDocs,
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
		Actions: actions,
	}

	testUtils.ExecuteTestCase(t, test)
}

// Every node merges the increments of the other four, which the two node case cannot cover.
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

	// Every pair, so an update reaches the others directly rather than being relayed.
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

	// Interleaved, as above.
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
		// As above.
		MultiplierExcludes: []string{
			multiplier.SecondaryIndex,
			multiplier.SignedDocs,
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
		Actions: actions,
	}

	testUtils.ExecuteTestCase(t, test)
}
