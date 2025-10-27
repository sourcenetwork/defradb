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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQuery_WithIndexOnNotMigratedDocs_ShouldNotHinder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @index
						points: Int
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "John",
					"points": 100,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Alice",
					"points": 200,
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreiaggrtq3p5esmkyqnmuh2dhwakhxmivacc6xj2vqaig566zc7mq6u",
					DestinationSchemaVersionID: "bafyreibppsfeecybqx2n24iuev2nzsti7zr4n6tkbeg4kcw5expy6lmgdm",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "points",
									"value": 50,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": int64(150),
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {name: {_eq: "John"}}) {
						name
						points
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithIndexOnMigratedField_ShouldUseIndexWithMigratedValues(t *testing.T) {
	const oldSchemaVersion = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
	const newSchemaVersion = "bafyreic75wgihcghabkb6idsnp2rrdugo6drwshiu6wnwypz2oyfwmqdeq"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String 
						age: Int @index
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      oldSchemaVersion,
					DestinationSchemaVersionID: newSchemaVersion,
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "age",
									"value": 5,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithIndexOnMigratedFieldAndSettingOldVersionAsActive_ShouldUseIndexWithOldValues(t *testing.T) {
	const oldSchemaVersion = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
	const newSchemaVersion = "bafyreic75wgihcghabkb6idsnp2rrdugo6drwshiu6wnwypz2oyfwmqdeq"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String 
						age: Int @index
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      oldSchemaVersion,
					DestinationSchemaVersionID: newSchemaVersion,
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "age",
									"value": 5,
								},
							},
						},
					},
				},
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: oldSchemaVersion,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithIndexAppliedAfterMigration_ShouldIndexDocsOnLatestVersion(t *testing.T) {
	const oldSchemaVersion = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
	const newSchemaVersion = "bafyreic75wgihcghabkb6idsnp2rrdugo6drwshiu6wnwypz2oyfwmqdeq"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String 
						age: Int 
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      oldSchemaVersion,
					DestinationSchemaVersionID: newSchemaVersion,
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "age",
									"value": 5,
								},
							},
						},
					},
				},
			},
			testUtils.CreateIndex{
				FieldName: "age",
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithIndexAppliedAfterSetActiveVersion_ShouldIndexDocsOnActiveVersion(t *testing.T) {
	const oldSchemaVersion = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
	const newSchemaVersion = "bafyreic75wgihcghabkb6idsnp2rrdugo6drwshiu6wnwypz2oyfwmqdeq"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String 
						age: Int 
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      oldSchemaVersion,
					DestinationSchemaVersionID: newSchemaVersion,
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "age",
									"value": 5,
								},
							},
						},
					},
				},
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: oldSchemaVersion,
			},
			testUtils.CreateIndex{
				FieldName: "age",
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
