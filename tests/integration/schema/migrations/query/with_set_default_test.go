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

func TestSchemaMigrationQuery_WithSetDefaultToLatest_AppliesForwardMigration(t *testing.T) {
	schemaVersionID2 := "bafkreibzqyjmyjs7vyo2q4h2tv5rbdbe4lv7tjbl5esilmobhgclia2juy"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreiadnck34zzbwayjw3aeubw7eg4jmgtwoibu35tkxbjpar5rzxkdpu",
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
			testUtils.SetDefaultSchemaVersion{
				SchemaVersionID: schemaVersionID2,
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
						"name":     "John",
						"verified": true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithSetDefaultToOriginal_AppliesInverseMigration(t *testing.T) {
	schemaVersionID1 := "bafkreiadnck34zzbwayjw3aeubw7eg4jmgtwoibu35tkxbjpar5rzxkdpu"
	schemaVersionID2 := "bafkreibzqyjmyjs7vyo2q4h2tv5rbdbe4lv7tjbl5esilmobhgclia2juy"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						verified: Boolean
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.SetDefaultSchemaVersion{
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
			testUtils.SetDefaultSchemaVersion{
				SchemaVersionID: schemaVersionID1,
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
						"name": "John",
						// The inverse lens migration has been applied, clearing the verified field
						"verified": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithSetDefaultToOriginalVersionThatDocWasCreatedAt_ClearsMigrations(t *testing.T) {
	schemaVersionID1 := "bafkreiadnck34zzbwayjw3aeubw7eg4jmgtwoibu35tkxbjpar5rzxkdpu"
	schemaVersionID2 := "bafkreibzqyjmyjs7vyo2q4h2tv5rbdbe4lv7tjbl5esilmobhgclia2juy"

	test := testUtils.TestCase{
		Description: "Test schema migration",
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(true),
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
			testUtils.SetDefaultSchemaVersion{
				SchemaVersionID: schemaVersionID1,
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
						"name": "John",
						// The inverse lens migration has not been applied, the document is returned as it was defined
						"verified": false,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
