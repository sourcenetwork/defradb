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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PPeerCreateWithNewFieldSyncsDocsToOlderSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will directly create a doc on
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
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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
			testUtils.Request{
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

func TestP2PPeerCreateWithNewFieldSyncsDocsToNewerSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will sync docs to
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
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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

func TestP2PPeerCreateWithNewFieldSyncsDocsToUpdatedSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Patch the schema on all nodes
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
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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
func TestP2PPeerCreateWithNewFieldDocSyncedBeforeReceivingNodeSchemaUpdatedDoesNotReturnNewField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
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
			testUtils.SchemaPatch{
				// Patch the schema on the first node only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.CreateDoc{
				// Create the doc with a value in the new field on the first node only, and allow the values to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.SchemaPatch{
				// Update the schema on the second node
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
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
			testUtils.Request{
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
