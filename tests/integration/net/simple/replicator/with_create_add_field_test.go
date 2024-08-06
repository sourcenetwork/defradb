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

func TestP2POneToOneReplicatorCreateWithNewFieldSyncsDocsToOlderSchemaVersion(t *testing.T) {
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
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

func TestP2POneToOneReplicatorCreateWithNewFieldSyncsDocsToNewerSchemaVersion(t *testing.T) {
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
				// Patch the schema on the node that we sync docs to
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
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

func TestP2POneToOneReplicatorCreateWithNewFieldSyncsDocsToUpdatedSchemaVersion(t *testing.T) {
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
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
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
