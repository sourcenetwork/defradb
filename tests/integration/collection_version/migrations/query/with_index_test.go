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

const (
	indexedSchemaV1 = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
	indexedSchemaV2 = "bafyreic75wgihcghabkb6idsnp2rrdugo6drwshiu6wnwypz2oyfwmqdeq"
	indexedSchemaV3 = "bafyreibz2wolrhx2vnrvyh7vg5rlehyekhnivthqom5zft5ku4mhpzi2fa"
	indexedSchemaV4 = "bafyreigpp4pjidheelridixgeppcki44t33dugru5b4hqnicyue5jqtudq"
	indexedSchemaV5 = "bafyreiclhvinnlruvathdewsz3inr55kh4koogikxds3akm23vzzv5vffa"
)

func TestSchemaMigrationQuery_WithIndexOnMigratedField_ShouldUseIndexWithMigratedValues(t *testing.T) {
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
					SourceSchemaVersionID:      indexedSchemaV1,
					DestinationSchemaVersionID: indexedSchemaV2,
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
					SourceSchemaVersionID:      indexedSchemaV1,
					DestinationSchemaVersionID: indexedSchemaV2,
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
				VersionID: indexedSchemaV1,
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
					SourceSchemaVersionID:      indexedSchemaV1,
					DestinationSchemaVersionID: indexedSchemaV2,
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
					SourceSchemaVersionID:      indexedSchemaV1,
					DestinationSchemaVersionID: indexedSchemaV2,
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
				VersionID: indexedSchemaV1,
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

// setupDistantVersions creates a chain of 5 versions with documents.
// v1 (age: Int @index) -> v2 (adds level) -> v3 (adds points) -> v4 (adds rank) -> v5 (adds score)
func setupDistantVersions() []any {
	return []any{
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
		testUtils.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }
				]
			`,
		},
		testUtils.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }
				]
			`,
		},
		testUtils.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "score", "Kind": "Int"} }
				]
			`,
		},
	}
}

// addMigrationBetweenV3AndV4 adds a lens migration between v3 and v4 that increments age by 5.
func addMigrationBetweenV3AndV4() any {
	return testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
			SourceSchemaVersionID:      indexedSchemaV3,
			DestinationSchemaVersionID: indexedSchemaV4,
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
	}
}

func TestSchemaMigrationQuery_SwitchToOldDistantVersionWithNoMigrations_ShouldUseOldIndexValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV1,
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

func TestSchemaMigrationQuery_SwitchToNewDistantVersionWithNoMigrations_ShouldUseOldIndexValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV1,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV5,
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

func TestSchemaMigrationQuery_SwitchToOldDistantVersionWithMigrationInBetween_ShouldReindexWithOldValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			addMigrationBetweenV3AndV4(),
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV1,
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

func TestSchemaMigrationQuery_SwitchToNewDistantVersionWithMigrationInBetween_ShouldReindexWithMigratedValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			addMigrationBetweenV3AndV4(),
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV1,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: indexedSchemaV5,
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
