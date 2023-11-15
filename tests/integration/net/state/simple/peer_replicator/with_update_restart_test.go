// Copyright 2022 Democratized Data Foundation
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

func TestP2PPeerReplicatorWithUpdateAndRestart(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
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
			// We need to wait and ensure that the create events are handled before
			// restarting the nodes as otherwise there is no gaurentee which side of
			// the restart that the create events are handled, resulting in flaky tests
			testUtils.WaitForSync{},
			testUtils.Restart{},
			testUtils.UpdateDoc{
				// Update John's Age on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Age": 60
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Age": int64(60),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
