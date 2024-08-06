// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PReplicatorUpdateWithNewFieldSyncsDocsToOlderSchemaVersionMultistep(t *testing.T) {
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will update the doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
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
				// We need to make sure any errors caused by the first update do not break the sync
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "Shahzad",
							"Email": "imnotyourbuddyguy@source.ca",
						},
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PReplicatorUpdateWithNewFieldSyncsDocsToOlderSchemaVersion(t *testing.T) {
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SchemaPatch{
				// Patch the schema on the node that we will directly update the doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "Shahzad",
							"Email": "imnotyourbuddyguy@source.ca",
						},
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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
