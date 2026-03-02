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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PAddDoesNotSync(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create Shahzad on all nodes
				Doc: `{
					"Name": "Shahzad",
					"Age": 300
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(300),
						},
						{
							"Age": int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(300),
						},
						// Peer sync should not sync new documents to nodes
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestP2PAddWithP2PCollection ensures that added documents reach the node that subscribes
// to the P2P collection topic but not the one that doesn't.
func TestP2PAddWithP2PCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				// Create Shahzad on all nodes
				Doc: `{
					"Name": "Shahzad",
					"Age": 30
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Addo",
					"Age": 28
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "Fred",
					"Age": 31
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(30),
						},
						{
							"Age": int64(21),
						},
						{
							"Age": int64(28),
						},
						// Peer sync should not sync new documents to nodes that is not subscribed
						// to the P2P collection.
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(31),
						},
						{
							"Age": int64(30),
						},
						{
							"Age": int64(21),
						},
						{
							"Age": int64(28),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PAdd_WithP2PCollectionWithNodeChain_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Having more than 3 nodes is important to test the robustness of the doc update message
			// processing function. Having more than 3 connected nodes means that there is a chance that
			// the message can be received multiple times simultaneously.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
                    type Users {
                        Name: String
                        Age: Int
                    }
                `,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 1,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 4,
				TargetNodeID: 3,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        2,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        3,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        4,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
                    "Name": "John",
                    "Age": 21
                }`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
                    Users {
                        Age
                    }
                }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PAdd_WithP2PCollectionOnLastNodeInNodeChain_ShouldPropagateUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
                    type Users {
                        Name: String
                        Age: Int
                    }
                `,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 1,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 3,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 4,
				TargetNodeID: 3,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        4,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
                    "Name": "John",
                    "Age": 21
                }`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(4),
				Request: `query {
                    Users {
                        Age
                    }
                }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PAdd_WithP2PCollectionAndSubscription_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
                    type Users {
                        Name: String
                        Age: Int
                    }
                `,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.SubscriptionRequest{
				NodeID: immutable.Some(1),
				Request: `subscription {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"Users": []map[string]any{
							{
								"Age": int64(21),
							},
						},
					},
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
                    "Name": "John",
                    "Age": 21
                }`,
			},
			testUtils.WaitForSync{},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
