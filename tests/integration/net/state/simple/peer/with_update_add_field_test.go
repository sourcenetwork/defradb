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

func TestP2PPeerUpdateWithNewFieldSyncsDocsToOlderSchemaVersionMultistep(t *testing.T) {
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
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will directly create a doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				// Update the new field on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.UpdateDoc{
				// Update the existing field on the first node only, and allow the value to sync
				// We need to make sure any errors caused by the first update to not break the sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shahzad"
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
				Results: []map[string]any{
					{
						"Name":  "Shahzad",
						"Email": "imnotyourbuddyguy@source.ca",
					},
				},
			},
			testUtils.Request{
				// The second update should still be received by the second node, updating Name
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PPeerUpdateWithNewFieldSyncsDocsToOlderSchemaVersion(t *testing.T) {
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
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will directly update the doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.UpdateDoc{
				// Update the new field and existing field on the first node only, and allow the values to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shahzad",
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
				Results: []map[string]any{
					{
						"Name":  "Shahzad",
						"Email": "imnotyourbuddyguy@source.ca",
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
