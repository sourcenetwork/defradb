// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
	"github.com/sourcenetwork/defradb/tests/integration/net/state/simple"
)

func TestP2PCreateDoesNotSync(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodePeers: map[int][]int{
			1: {
				0,
			},
		},
		SeedDocuments: map[int]map[int]string{
			0: {
				0: `{
					"Name": "Shahzad",
					"Age": 300
				}`,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					1: `{
						"Name": "John",
						"Age": 21
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Age": uint64(300),
				},
				1: {
					"Age": uint64(21),
				},
			},
			1: {
				0: {
					"Age": uint64(300),
				},
				// Peer sync should not sync new documents to nodes
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
