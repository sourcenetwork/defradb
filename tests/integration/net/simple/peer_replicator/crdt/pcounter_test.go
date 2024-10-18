// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_replicator_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PPeerReplicatorWithCreate_PCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
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
				Doc: `{
					"name": "John",
					"points": 0
				}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Shahzad",
					"points": 3000
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"points": int64(0),
						},
						{
							"points": int64(3000),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"points": int64(0),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(2),
				Request: `query {
					Users {
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"points": int64(0),
						},
						{
							"points": int64(3000),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PPeerReplicatorWithUpdate_PCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
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
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.UpdateDoc{
				// Update John's points on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
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
