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

// The parent-child distinction in these tests is as much documentation and test
// of the test system as of production.  See it as a santity check of sorts.
func TestP2PWithSingleDocumentSingleUpdateFromChild(t *testing.T) {
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

// The parent-child distinction in these tests is as much documentation and test
// of the test system as of production.  See it as a santity check of sorts.
func TestP2PWithSingleDocumentSingleUpdateFromParent(t *testing.T) {
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
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

// TestP2PWithSingleDocumentUpdatePerNode tests document syncing between two nodes with a single update per node
func TestP2PWithSingleDocumentUpdatePerNode(t *testing.T) {
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
			1: {
				0: {
					0: {
						`{
							"Age": 45
						}`,
					},
				},
			},
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
					"Age": testUtils.AnyOf{uint64(45), uint64(60)},
				},
			},
			1: {
				0: {
					"Age": testUtils.AnyOf{uint64(45), uint64(60)},
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2PWithSingleDocumentSingleUpdateDoesNotSyncToNonPeerNode(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// This last node is not marked for peer sync
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
			2: {
				// Update should not be synced to this node
				0: {
					"Age": uint64(21),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}

func TestP2PWithSingleDocumentSingleUpdateDoesNotSyncFromUnmappedNode(t *testing.T) {
	test := testUtils.P2PTestCase{
		NodeConfig: []*config.Config{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// This node is unmapped, updates applied to this node should
			// not be synced to the other nodes.
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
			2: {
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
					// Update should not be synced to this node
					"Age": uint64(21),
				},
			},
			1: {
				0: {
					// Update should not be synced to this node
					"Age": uint64(21),
				},
			},
			2: {
				0: {
					"Age": uint64(60),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}

// This test should cover the same production code as the above
// `TestP2PWithSingleDocumentSingleUpdateDoesNotSyncFromUnmappedNode` test, however
// it provides an additional sanity check for the somewhat complex test framework
// to ensure that the test code is functioning correctly here.
func TestP2PWithSingleDocumentSingleUpdateDoesNotSyncFromNonPeerNode(t *testing.T) {
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
			// Peer node is declared, but not mapped to the others. Updates applied
			// to this node should not be synced to the other nodes.
			2: {},
		},
		SeedDocuments: map[int]map[int]string{
			0: {
				0: `{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Updates: map[int]map[int]map[int][]string{
			2: {
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
					"Age": uint64(21),
				},
			},
			1: {
				0: {
					"Age": uint64(21),
				},
			},
			2: {
				// Update should not be synced to this node
				0: {
					"Age": uint64(60),
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}

// TestP2PWithMultipleDocumentUpdatesPerNode tests document syncing between two nodes with multiple updates per node.
func TestP2PWithMultipleDocumentUpdatesPerNode(t *testing.T) {
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
						`{
							"Age": 61
						}`,
						`{
							"Age": 62
						}`,
					},
				},
			},
			1: {
				0: {
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
		},
		Results: map[int]map[int]map[string]any{
			0: {
				0: {
					"Age": testUtils.AnyOf{uint64(47), uint64(62)},
				},
			},
			1: {
				0: {
					"Age": testUtils.AnyOf{uint64(47), uint64(62)},
				},
			},
		},
	}

	simple.ExecuteTestCase(t, test)
}
