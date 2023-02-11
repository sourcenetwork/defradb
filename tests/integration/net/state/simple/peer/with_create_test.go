// Copyright 2022 Democratized Data Foundation
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

// TestP2PCreateWithP2PCollection ensures that created documents reach the node that subscribes
// to the P2P collection topic but not the one that doesn't.
func TestP2PCreateWithP2PCollection(t *testing.T) {
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
		NodeP2PCollection: map[int][]int{
			1: {
				0,
			},
		},
		SeedDocuments: map[int]map[int]string{
			0: {
				0: `{
					"Name": "Shahzad",
					"Age": 30
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
					2: `{
						"Name": "Addo",
						"Age": 28
					}`,
				},
			},
			1: {
				0: {
					3: `{
						"Name": "Fred",
						"Age": 31
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Age": uint64(30),
				},
				1: {
					"Age": uint64(21),
				},
				2: {
					"Age": uint64(28),
				},
				// Peer sync should not sync new documents to nodes that is not subscribed
				// to the P2P collection.
			},
			1: {
				0: {
					"Age": uint64(30),
				},
				1: {
					"Age": uint64(21),
				},
				2: {
					"Age": uint64(28),
				},
				3: {
					"Age": uint64(31),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
