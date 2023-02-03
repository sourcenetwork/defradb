// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer

import (
	"testing"

	"github.com/sourcenetwork/defradb/config"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
	"github.com/sourcenetwork/defradb/tests/integration/net/state/one_to_many"
)

// This test asserts that relational documents do not fail to sync if their related
// document does not exist at the destination.
func TestP2POneToManyPeerWithCreateUpdateLinkingSyncedDocToUnsyncedDoc(t *testing.T) {
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
			1: {
				0: `{
					"Name": "Gulistan"
				}`,
			},
		},
		Creates: map[int]map[int]map[int]string{
			0: {
				0: {
					// NodePeers do not sync new documents so this will not be synced
					// to node 1.
					1: `{
						"Name": "Saadi"
					}`,
				},
			},
		},
		Updates: map[int]map[int]map[int][]string{
			0: {
				1: {
					0: {
						`{
							"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						}`,
					},
				},
			},
		},
		Results: map[int]map[int]map[string]any{
			0: {
				1: {
					"Name": "Saadi",
				},
				0: {
					"Name":      "Gulistan",
					"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
				},
			},
			1: {
				0: {
					"Name":      "Gulistan",
					"Author_id": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
				},
				// "Saadi" was not synced to node 1, the update did not
				// result in an error and synced to relational id even though "Saadi"
				// does not exist in this node.
			},
		},
	}

	one_to_many.ExecuteTestCase(t, test)
}
