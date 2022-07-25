// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"testing"

	"github.com/sourcenetwork/defradb/config"
)

// TestP2PWithSingleDocumentUpdatePerNode tests document syncing between two nodes with a single update per node
func TestP2PWithSingleDocumentUpdatePerNode(t *testing.T) {
	test := P2PTestCase{
		NodeConfig: []*config.Config{
			randomNetworkingConfig(),
			randomNetworkingConfig(),
		},
		NodePeers: map[int][]int{
			1: {
				0,
			},
		},
		SeedDocuments: []string{
			`{
				"Name": "John",
				"Age": 21
			}`,
		},
		Updates: map[int]map[int][]string{
			1: {
				0: {
					`{
						"Age": 45
					}`,
				},
			},
			0: {
				0: {
					`{
						"Age": 60
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]interface{}{
			0: {
				0: {
					"Age": uint64(45),
				},
			},
			1: {
				0: {
					"Age": uint64(60),
				},
			},
		},
	}

	executeTestCase(t, test)
}

// TestP2PWithMultipleDocumentUpdatesPerNode tests document syncing between two nodes with multiple updates per node.
func TestP2PWithMultipleDocumentUpdatesPerNode(t *testing.T) {
	test := P2PTestCase{
		NodeConfig: []*config.Config{
			randomNetworkingConfig(),
			randomNetworkingConfig(),
		},
		NodePeers: map[int][]int{
			1: {
				0,
			},
		},
		SeedDocuments: []string{
			`{
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
					`{
						"Age": 61
					}`,
					`{
						"Age": 62
					}`,
				},
			},
			1: {
				0: {
					`{
						"Age": 45
					}`,
					`{
						"Age": 46
					}`,
					`{
						"Age": 47
					}`,
				},
			},
		},
		Results: map[int]map[int]map[string]interface{}{
			0: {
				0: {
					"Age": uint64(47),
				},
			},
			1: {
				0: {
					"Age": uint64(62),
				},
			},
		},
	}

	executeTestCase(t, test)
}
