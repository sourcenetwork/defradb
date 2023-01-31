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

func TestP2PWithSingleDocumentDelete(t *testing.T) {
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
					"Age": 30
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
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2PWithMultipleDocumentDelete(t *testing.T) {
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
					"Age": 30
				}`,
				1: `{
					"Name": "Fred",
					"Age": 22
				}`,
				2: `{
					"Name": "John",
					"Age": 31
				}`,
			},
		},
		Deletes: map[int]map[int][]int{
			0: {
				0: {
					0,
					1,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				2: {
					"Name": "John",
					"Age":  uint64(31),
				},
			},
			1: {
				2: {
					"Name": "John",
					"Age":  uint64(31),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
