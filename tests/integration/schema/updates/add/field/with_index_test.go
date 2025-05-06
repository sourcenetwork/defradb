// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldSimple_WithExistingIndexDocsCreatedAfterPatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test patching schema for collection with index still works",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			// It is important to test that the index shows up in both the `GetIndexes` call,
			// *and* the `GetCollections` call, as indexes are stored in multiple places and we had a bug
			// where patching a schema would result in the index disappearing from one of those locations.
			testUtils.GetIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Users_name_ASC",
						ID:     1,
						Unique: false,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
				},
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						Indexes: []client.IndexDescription{
							{
								Name:   "Users_name_ASC",
								ID:     1,
								Unique: false,
								Fields: []client.IndexedFieldDescription{
									{
										Name: "name",
									},
								},
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(1).WithIndexFetches(1),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldSimple_WithExistingIndexDocsCreatedBeforePatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test patching schema for collection with index still works",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
					}
				`,
			},
			// It is important to test this with docs created *before* the patch, as well as after (see other test).
			// A bug was missed by missing this test case.
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			// It is important to test that the index shows up in both the `GetIndexes` call,
			// *and* the `GetCollections` call, as indexes are stored in multiple places and we had a bug
			// where patching a schema would result in the index disappearing from one of those locations.
			testUtils.GetIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Users_name_ASC",
						ID:     1,
						Unique: false,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
				},
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						Indexes: []client.IndexDescription{
							{
								Name:   "Users_name_ASC",
								ID:     1,
								Unique: false,
								Fields: []client.IndexedFieldDescription{
									{
										Name: "name",
									},
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(1).WithIndexFetches(1),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
