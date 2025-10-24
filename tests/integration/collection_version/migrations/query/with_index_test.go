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

func TestSchemaMigrationQueryWithIndexOnMigratedDocs(t *testing.T) {
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
					DestinationSchemaVersionID: "bafyreialhfulxmzcbgrqkuw76icaxb25yplnse2qb5vhufbiyuge3ifvuu",
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

// TODO: after migration
func TestSchemaMigrationQueryWithIndexCreatedBeforeMigration(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						username: String @index
						score: Int
					}
				`,
			},
			// Create documents before migration
			testUtils.CreateDoc{
				Doc: `{
					"username": "player1",
					"score": 500
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"username": "player2",
					"score": 750
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"username": "player3",
					"score": 300
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "rank", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafyreigb4bsjwp5wp4pylyigecdqu3vvgqu7t7iwx754fvnma4jqrkdn4q",
					DestinationSchemaVersionID: "bafyreigzrsd7x5h4eqxwvzbacgzlnhrwqw5yqfuf2sfxh6rrvjmsl4vlcu",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "score",
									"value": 100,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {username: {_eq: "player2"}}) {
						username
						score
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(1).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {username: {_eq: "player2"}}) {
						username
						score
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"username": "player2",
							"score":    int64(850),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TODO: temp tests, revisit
func TestSchemaMigrationQueryWithIndexOnNonMigratedField2(t *testing.T) {
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
					"name": "Bob",
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
					SourceSchemaVersionID:      "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma",
					DestinationSchemaVersionID: "bafyreihqxgaliyhnybhzu6373x3rrfqj2n63ipykol3x2qdi6djvigftdq",
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
