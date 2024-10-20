// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replicator

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2POneToOneReplicatorUpdate_PNCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			testUtils.CreateDoc{
				// This document is created in first node before the replicator is set up.
				// Updates should be synced across nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
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
