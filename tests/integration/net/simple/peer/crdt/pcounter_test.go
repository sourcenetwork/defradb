// Copyright 2024 Democratized Data Foundation
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

func TestP2PUpdate_WithPCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.CreateDoc{
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
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
				Doc: `{
					"points": 10
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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

func TestP2PUpdate_WithPCounterSimultaneousUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.CreateDoc{
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
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Age": 45
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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
