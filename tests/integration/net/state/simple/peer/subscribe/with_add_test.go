// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscribe_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestP2PSubscribeAddSingle ensures that created documents reach the node that subscribes
// to the P2P collection topic but not the one that doesn't.
func TestP2PSubscribeAddSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
					// Peer sync should not sync new documents to nodes that is not subscribed
					// to the P2P collection.
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
					type Giraffes {
						name: String
					}
					type Bears {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 2,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				Doc: `{
					"name": "Gillian"
				}`,
			},
			testUtils.CreateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 2,
				Doc: `{
					"name": "Bjorn"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// John the User has been synced.
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
			testUtils.Request{
				// Gillian the Giraffe has not been synced, as the collection (1)
				// was not subscribed to.
				NodeID: immutable.Some(1),
				Request: `query {
					Giraffes {
						name
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.Request{
				// Bjorn the Bear has been synced.
				Request: `query {
					Bears {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Bjorn",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddSingleErroneousCollectionID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionID:  testUtils.NonExistentCollectionID,
				ExpectedError: "datastore: key not found",
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// Nothing should sync
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddValidThenErroneousCollectionID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionID:  testUtils.NonExistentCollectionID,
				ExpectedError: "datastore: key not found",
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// The subscription for collection 0 should still be active
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddNone(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
