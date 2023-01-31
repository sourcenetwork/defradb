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

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
	"github.com/sourcenetwork/defradb/tests/integration/net/state/simple"
)

// TestP2FullPReplicator tests document syncing between a node and a replicator.
func TestP2POneToOneReplicatorWithDelete(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodeReplicators: map[int][]int{
			0: {
				1,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					0: `{
						"Name": "John",
						"Age": 21
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
			0: {},
			1: {},
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorManyDocsWithDelete(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodeReplicators: map[int][]int{
			0: {
				1,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					0: `{
						"Name": "John",
						"Age": 21
					}`,
					1: `{
						"Name": "Fred",
						"Age": 22
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
					"Age": uint64(22),
				},
			},
			1: {
				1: {
					"Age": uint64(22),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2POneToManyReplicatorManyDocsWithDelete(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
		},
		NodeReplicators: map[int][]int{
			0: {
				1,
				2,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					0: `{
						"Name": "John",
						"Age": 21
					}`,
					1: `{
						"Name": "Fred",
						"Age": 22
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
					"Age": uint64(22),
				},
			},
			1: {
				1: {
					"Age": uint64(22),
				},
			},
			2: {
				1: {
					"Age": uint64(22),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
