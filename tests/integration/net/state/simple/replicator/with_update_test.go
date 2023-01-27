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

func TestP2POneToOneReplicatorUpdatesDocCreatedBeforeReplicatorConfig(t *testing.T) {
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
		SeedDocuments: map[int]map[int]string{
			// This document is created in all nodes before the replicator is set up.
			// Updates should be synced across nodes.
			0: {
				0: `{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
			0: {
				0: {
					0: {
						`{
							"Age": 60
						}`,
					},
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
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorUpdatesDocCreatedBeforeReplicatorConfigWithNodesInversed(t *testing.T) {
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
		SeedDocuments: map[int]map[int]string{
			// This document is created in all nodes before the replicator is set up.
			// Updates should be synced across nodes.
			0: {
				0: `{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
			// Note: The update is applied to the target node (not source) specified in the config.
			1: {
				0: {
					0: {
						`{
						"Age": 60
					}`,
					},
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
		},
	}

	simple.ExecuteTestCase(t, test)
}
