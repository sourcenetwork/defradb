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

func TestP2PSubscribeAddAndRemoveSingle(t *testing.T) {
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
				CollectionIDs: []int{0},
			},
			testUtils.UnsubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
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
				// John has not been synced, as it was removed from the subscription set
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
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddAndRemoveMultiple(t *testing.T) {
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
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0, 1},
			},
			testUtils.UnsubscribeToCollection{
				NodeID: 1,
				// Unsubscribe from Users, but remain subscribed to Giraffes
				CollectionIDs: []int{0},
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
			testUtils.WaitForSync{},
			testUtils.Request{
				// John the User has not been synced, as Users was removed from the subscription set.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.Request{
				// Gillian the Giraffe has still been synced, as it was not removed from the subscription set.
				NodeID: immutable.Some(1),
				Request: `query {
					Giraffes {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Gillian",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddSingleAndRemoveErroneous(t *testing.T) {
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
				CollectionIDs: []int{0},
			},
			testUtils.UnsubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0, testUtils.NonExistentCollectionID},
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
				// John has been synced, as the unsubscribe errored and should not have affected
				// the subscription to collection 0.
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

func TestP2PSubscribeAddSingleAndRemoveNone(t *testing.T) {
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
				CollectionIDs: []int{0},
			},
			testUtils.UnsubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				// John has been synced, as nothing was removed from the subscription set
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
