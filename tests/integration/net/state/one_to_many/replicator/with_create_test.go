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

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
	"github.com/sourcenetwork/defradb/tests/integration/net/state/one_to_many"
)

// TestP2FullPReplicator tests document syncing between a node and a replicator.
func TestP2POneToManyReplicator(t *testing.T) {
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
						"Name": "Saadi"
					}`,
				},
				1: {
					1: `{
						"Name": "Gulistan",
						"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Name": "Saadi",
				},
				1: {
					"Name":      "Gulistan",
					"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
				},
			},
			1: {
				0: {
					"Name": "Saadi",
				},
				1: {
					"Name":      "Gulistan",
					"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
				},
			},
		},
	}

	one_to_many.ExecuteTestCase(t, test)
}
