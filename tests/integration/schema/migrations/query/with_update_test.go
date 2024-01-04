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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryWithUpdateRequest(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, with update request",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreig3zt63qt7bkji47etyu2sqtzroa3tcfdxgwqc3ka2ijy63refq3a",
					DestinationSchemaVersionID: "bafkreia4m6sn2rfypj2velvwpyude22fcb5jyfzum2eh3cdzg4a3myj5nu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(data: "{\"name\":\"Johnnnn\"}") {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Johnnnn",
						// We need to assert that the migration has been run within the context
						// of the update
						"verified": true,
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Johnnnn",
						// We need to assert that the effects of the migration executed within the
						// update have been persisted
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQueryWithMigrationRegisteredAfterUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, with migration registered after update",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.UpdateDoc{
				// Update the document **before** registering the migration
				Doc: `{
					"name":	"Johnnnn"
				}`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreig3zt63qt7bkji47etyu2sqtzroa3tcfdxgwqc3ka2ijy63refq3a",
					DestinationSchemaVersionID: "bafkreia4m6sn2rfypj2velvwpyude22fcb5jyfzum2eh3cdzg4a3myj5nu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Johnnnn",
						// As the document was updated before the migration was registered
						// the migration will not have been run
						"verified": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
