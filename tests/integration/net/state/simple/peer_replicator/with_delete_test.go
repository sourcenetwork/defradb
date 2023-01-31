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
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
	"github.com/sourcenetwork/defradb/tests/integration/net/state/simple"
)

func TestP2PPeerReplicatorWithDelete(t *testing.T) {
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
		SeedDocuments: map[int]map[int]string{
			0: {
				0: `{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Deletes: map[int]map[int][]int{
			0: {
				0: {
					0,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {},
			1: {},
			2: {},
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2PPeerReplicatorWithCreateThenDelete(t *testing.T) {
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
		SeedDocuments: map[int]map[int]string{
			0: {
				0: `{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					1: `{
						"Name": "Shahzad",
						"Age": 30
					}`,
				},
			},
		},
		Deletes: map[int]map[int][]int{
			0: {
				0: {
					0,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				1: {
					"Name": "Shahzad",
					"Age":  uint64(30),
				},
			},
			1: {},
			2: {
				1: {
					"Name": "Shahzad",
					"Age":  uint64(30),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
