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

func TestSchemaMigrationQuery_WithFilter_ShouldFilterFMigration(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }]`,
				Lens: immutable.Some(
					model.Lens{
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
				),
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
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithFilterAndMigrationBetweenOldVersions_ShouldApplyMigration(t *testing.T) {
	const (
		schemaV3 = "bafyreidmsarf4ac4eihxk3ocqfort3e3pxhb7eumatvkanjsxxkjrn3h6a"
		schemaV4 = "bafyreidptieeo3tckkyi6jnomavo3noy2mxuv7dfuc76pf2vgxm6ilfazq"
	)

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }]`,
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }]`,
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }]`,
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "score", "Kind": "Int"} }]`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      schemaV3,
					DestinationCollectionVersionID: schemaV4,
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
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaMigrationQuery_WithFilterAndMigrationInOldPatch_ShouldApplyMigration2(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  20,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Fred",
					"age":  25,
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Islam",
					"age":  32,
				},
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }]`,
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "points", "Kind": "Int"} }]`,
			},
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }]`,
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
			testUtils.PatchCollection{
				Patch: `[{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "score", "Kind": "Int"} }]`,
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
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
