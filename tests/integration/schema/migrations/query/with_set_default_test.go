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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQuery_WithSetDefaultToLatest_AppliesForwardMigration(t *testing.T) {
	schemaVersionID2 := "bafyreiecoh66au3ofetd7lyetuqm7xlsj3ln777l7sw324jsutel3w6u5m"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "verified",
								"value": true,
							},
						},
					},
				}),
			},
			testUtils.SetActiveCollectionVersion{
				SchemaVersionID: schemaVersionID2,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"verified": true,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithSetDefaultToOriginal_AppliesInverseMigration(t *testing.T) {
	schemaVersionID1 := "bafyreibvat5vc4nxjbxlewe7y2gpskfugypzqbqwal2zwaa7bnuiedyuyy"
	schemaVersionID2 := "bafyreiecoh66au3ofetd7lyetuqm7xlsj3ln777l7sw324jsutel3w6u5m"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				SchemaVersionID: schemaVersionID2,
			},
			// Create John using the new schema version
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"verified": true
				}`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      schemaVersionID1,
					DestinationSchemaVersionID: schemaVersionID2,
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
			// Set the schema version back to the original
			testUtils.SetActiveCollectionVersion{
				SchemaVersionID: schemaVersionID1,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// The inverse lens migration has been applied, clearing the verified field
							"verified": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithSetDefaultToOriginalVersionThatDocWasCreatedAt_ClearsMigrations(t *testing.T) {
	schemaVersionID1 := "bafyreibvat5vc4nxjbxlewe7y2gpskfugypzqbqwal2zwaa7bnuiedyuyy"
	schemaVersionID2 := "bafyreiecoh66au3ofetd7lyetuqm7xlsj3ln777l7sw324jsutel3w6u5m"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			// Create John using the original schema version
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"verified": false
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      schemaVersionID1,
					DestinationSchemaVersionID: schemaVersionID2,
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
			// Set the schema version back to the original
			testUtils.SetActiveCollectionVersion{
				SchemaVersionID: schemaVersionID1,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// The inverse lens migration has not been applied, the document is returned as it was defined
							"verified": false,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
