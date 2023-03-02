// Copyright 2022 Democratized Data Foundation
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

func TestP2POneToOneReplicatorWithCreateWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// This document is created in node `0` after the replicator has
				// been set up. Its creation and future updates should be synced
				// across all configured nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
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
						"Age": uint64(60),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
