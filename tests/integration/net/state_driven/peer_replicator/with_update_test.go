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

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state_driven"
)

// This test documents a bug and the behaviour should be corrected
// https://github.com/sourcenetwork/defradb/issues/1000
func TestP2PPeerReplicatorWithUpdate(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodePeers: map[int][]int{
			1: {
				0,
			},
		},
		NodeReplicators: map[int][]int{
			0: {
				2,
			},
		},
		SeedDocuments: map[int]string{
			0: `{
				"Name": "John",
				"Age": 21
			}`,
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"Age": 60
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Age": uint64(60),
				},
			},
			1: {
				0: {
					"Age": uint64(60),
				},
			},
			2: {
				0: {
					// This is incorrect behaviour - node 2 should not
					// be updated and this value should be `21`
					"Age": uint64(60),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
