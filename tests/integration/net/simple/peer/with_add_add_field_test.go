// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package peer_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PPeerAddWithNewFieldSyncsDocsToOlderCollectionVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.PatchCollection{
				// Patch the collection on the node that we will directly add a doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
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
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "John",
							"Email": "imnotyourbuddyguy@source.ca",
						},
					},
				},
			},
			&action.Request{
				// John should still be synced to the second node, even though it has
				// not been updated to contain the new 'Email' field.
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PPeerAddWithNewFieldSyncsDocsToNewerCollectionVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.PatchCollection{
				// Patch the collection on the node that we will sync docs to
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
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
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				// John should still be synced to the second node, even though it has
				// been updated with a new 'Email' field that does not exist on the
				// source node.
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PPeerAddWithNewFieldSyncsDocsToUpdatedCollectionVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.PatchCollection{
				// Patch the collection on all nodes
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
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
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "John",
							"Email": "imnotyourbuddyguy@source.ca",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents unwanted behaviour and should be changed when
// https://github.com/sourcenetwork/defradb/issues/2255 is fixed.
func TestP2PPeerAddWithNewFieldDocSyncedBeforeReceivingNodeSchemaUpdatedDoesNotReturnNewField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
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
			&action.PatchCollection{
				// Patch the collection on the first node only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			&action.AddDoc{
				// Add the doc with a value in the new field on the first node only, and allow the values to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			&action.PatchCollection{
				// Update the collection on the second node
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users {
							Name
							Email
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "John",
							"Email": "imnotyourbuddyguy@source.ca",
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `
					query {
						Users {
							Name
							Email
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							// The email should be returned but it is not
							"Email": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
