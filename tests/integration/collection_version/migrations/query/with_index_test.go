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

const (
	colVersionV1 = "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq"
	colVersionV2 = "bafyreighc6zz7674lpd3vwbd3bve5elzol3ijntwtzmw6cspnxkfijdsxa"
	colVersionV3 = "bafyreidmsarf4ac4eihxk3ocqfort3e3pxhb7eumatvkanjsxxkjrn3h6a"
	colVersionV4 = "bafyreidptieeo3tckkyi6jnomavo3noy2mxuv7dfuc76pf2vgxm6ilfazq"
	colVersionV5 = "bafyreia2ls3vfvwbgaunr5si5cpo3be5m7vtbmlzxuzvls5laz74zpwrg4"
)

func TestCollectionMigrationQuery_WithIndexOnNotMigratedDocs_ShouldNotHinder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  40,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
			&action.Request{
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(35),
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {name: {_eq: "John"}}) {
						name
						age
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQuery_WithIndexOnMigratedField_ShouldUseIndexWithMigratedValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_WithIndexOnMigratedFieldAndSettingOldVersionAsActive_ShouldUseIndexWithOldValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
				VersionID: colVersionV1,
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_WithIndexAppliedAfterMigration_ShouldIndexDocsOnLatestVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int 
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
			&action.NewIndex{
				FieldName: "age",
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_WithIndexAppliedAfterSetActiveVersion_ShouldIndexDocsOnActiveVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int 
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
				VersionID: colVersionV1,
			},
			&action.NewIndex{
				FieldName: "age",
			},
			&action.Request{
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
			&action.Request{
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
		&action.AddCollection{
			SDL: `
				type Users {
					name: String
					age: Int @index
				}
			`,
		},
		&action.AddDoc{
			DocMap: map[string]any{
				"name": "Andy",
				"age":  20,
			},
		},
		&action.AddDoc{
			DocMap: map[string]any{
				"name": "John",
				"age":  30,
			},
		},
		&action.AddDoc{
			DocMap: map[string]any{
				"name": "Fred",
				"age":  25,
			},
		},
		&action.AddDoc{
			DocMap: map[string]any{
				"name": "Islam",
				"age":  32,
			},
		},
		&action.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
				]
			`,
		},
		&action.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }
				]
			`,
		},
		&action.PatchCollection{
			Patch: `
				[
					{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }
				]
			`,
		},
		&action.PatchCollection{
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
			SourceCollectionVersionID:      colVersionV3,
			DestinationCollectionVersionID: colVersionV4,
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

// We don't have a way to test if reindexing really happened, but we can check if system behaves as expected.
func TestCollectionMigrationQuery_SwitchToOldDistantVersionWithNoMigrations_ShouldNotReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			&action.Request{
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
			&action.Request{
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

// We don't have a way to test if reindexing really happened, but we can check if system behaves as expected.
func TestCollectionMigrationQuery_SwitchToNewDistantVersionWithNoMigrations_ShouldNotReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV5,
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_SwitchToOldDistantVersionWithMigrationInBetween_ShouldReindexWithOldValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			addMigrationBetweenV3AndV4(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_SwitchToNewDistantVersionWithMigrationInBetween_ShouldReindexWithMigratedValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			addMigrationBetweenV3AndV4(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV5,
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_ApplyingMigrationBetweenOldVersions_ShouldReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV5,
			},
			addMigrationBetweenV3AndV4(),
			&action.Request{
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
			&action.Request{
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

// We don't have a way to test if reindexing really happened, but we can check if system behaves as expected.
func TestCollectionMigrationQuery_ApplyingMigrationBetweenNewVersions_ShouldNotReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			addMigrationBetweenV3AndV4(),
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_ApplyingMigrationToUnknownVersionsThenPatch_ShouldReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      colVersionV1,
					DestinationCollectionVersionID: colVersionV2,
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
			// Now patch to actually create v2 - this should trigger reindexing
			// even though the patch itself doesn't touch indexes
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_ApplyingMigrationWithPatching_ShouldReindex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.IncrementModulePath,
							Arguments: map[string]any{
								"field": "age",
								"value": 5,
							},
						},
					},
				}),
			},
			&action.Request{
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
			&action.Request{
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

func TestCollectionMigrationQuery_WithBranchedVersionsAndMigration_ShouldApplyMigrationCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Alice",
					"age":  25,
				},
			},
			// Create branch A: v1 -> v2 (no migration)
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			// Create branch B: v1 -> v3 (with migration: age+5)
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.IncrementModulePath,
							Arguments: map[string]any{
								"field": "age",
								"value": 5,
							},
						},
					},
				}),
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(35),
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV2,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(30),
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(30),
						},
					},
				},
			},
			// Switch back to branch B (v3 with migration) - should reindex with migrated values
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV3,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(35),
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionMigrationQuery_WithThreeBranchedVersions_ShouldApplyCorrectMigrationPerBranch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  20,
				},
			},
			// Create branch A: v1 -> v2 (no migration)
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			// Switch back to v1 to create branch B
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			// Create branch B: v1 -> v3 (migration: age+5)
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.IncrementModulePath,
							Arguments: map[string]any{
								"field": "age",
								"value": 5,
							},
						},
					},
				}),
			},
			// Switch back to v1 to create branch C
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			// Create branch C: v1 -> v4 (migration: age+10)
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.IncrementModulePath,
							Arguments: map[string]any{
								"field": "age",
								"value": 10,
							},
						},
					},
				}),
			},
			// Test branch C (v4): age should be 30 (20 + 10)
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(30), // 20 + 10
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			// Switch to branch B (v3): age should be 25 (20 + 5)
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV3,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 25}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(25), // 20 + 5
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 25}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			// Switch to branch A (v2): age should be 20 (no migration)
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV2,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 20}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(20), // original
						},
					},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 20}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			// Switch back to root (v1): age should be 20 (original)
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV1,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 20}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(20), // original
						},
					},
				},
			},
			// Final switch back to branch C (v4): verify age is 30 again
			testUtils.SetActiveCollectionVersion{
				VersionID: colVersionV4,
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(30), // 20 + 10
						},
					},
				},
			},
			&action.Request{
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
