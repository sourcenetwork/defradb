// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryWithP2PReplicatedDocOnOtherSchemaBranch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.SchemaPatch{
				// Patch first node only
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				// Register the migration on both nodes.
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreihax57fohcdupqr2l4heoqxdsiggjfeaubr44tgrz4xqdgvnid4xy",
					DestinationSchemaVersionID: "bafkreifer354qmdrwdtae5n3k7sbl2oauis3mz24fk46p3otub7ojznobq",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "name",
									"value": "Fred",
								},
							},
						},
					},
				},
			},
			testUtils.SchemaPatch{
				// Patch second node with different patch
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "phone",
								"value": "1234567890",
							},
						},
					},
				}),
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.CreateDoc{
				// Create John on the first (source) node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name":     "John",
						"verified": true,
					},
				},
			},
			testUtils.Request{
				// Node 1 should yield results migrated down to schema version 1, then up to schema version 3.
				NodeID: immutable.Some(1),
				Request: `
					query {
						Users {
							name
							phone
						}
					}
				`,
				Results: []map[string]any{
					{
						// name has been cleared by the inverse of the migration from version 1 to 2
						"name": nil,
						// phone has been set by the migration from version 1 to 3
						"phone":    "1234567890",
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
