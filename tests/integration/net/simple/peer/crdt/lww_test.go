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

func TestP2PUpdate_WithLWWConcurrentDifferentFields_BothFieldsPreserved(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
						score: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21,
					"score": 0
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
					"name": "Jane"
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"age": 45
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(2),
				Doc: `{
					"score": 100
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						name
						age
						score
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Jane",
							"age":   int64(45),
							"score": int64(100),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PUpdate_WithLWWConcurrentSameField_ConvergesToSameValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
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
					"age": 30
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"age": 45
				}`,
			},
			&action.UpdateDoc{
				NodeID: immutable.Some(2),
				Doc: `{
					"age": 60
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age": int64(60),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
