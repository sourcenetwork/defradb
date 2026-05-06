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

func TestP2PUpdate_WithPCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		//
		// Signed counter deltas are double-applied across peers.
		// https://github.com/sourcenetwork/defradb/issues/4742
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.SignedDocs},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				// Create Shahzad on all nodes
				Doc: `{
					"name": "Shahzad",
					"points": 10
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"points": 10
				}`,
			},
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
							"points": int64(20),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PUpdate_WithPCounterThreeNodeSimultaneousUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 0
				}`,
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
				SourceNodeID: 1,
				TargetNodeID: 2,
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
			testUtils.AddDocumentSubscription{
				NodeID: 2,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"points": 10
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"points": 20
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(2),
				Doc: `{
					"points": 30
				}`,
			},
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
							"points": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PUpdate_WithPCounterSimultaneousUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 0
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
			&action.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 45
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(90),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
