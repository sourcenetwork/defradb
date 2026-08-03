// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package field

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldSimple_WithExistingIndexDocsAddedAfterPatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			// It is important to test that the index shows up in both the `ListIndexes` call,
			// *and* the `GetCollections` call, as indexes are stored in multiple places and we had a bug
			// where patching a collection would result in the index disappearing from one of those locations.
			&action.ListIndexes{
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
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
			&action.Request{
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

func TestCollectionVersionUpdatesAddFieldSimple_WithExistingIndexDocsAddedBeforePatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index
					}
				`,
			},
			// It is important to test this with docs created *before* the patch, as well as after (see other test).
			// A bug was missed by missing this test case.
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			// It is important to test that the index shows up in both the `ListIndexes` call,
			// *and* the `GetCollections` call, as indexes are stored in multiple places and we had a bug
			// where patching a collection would result in the index disappearing from one of those locations.
			&action.ListIndexes{
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
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
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
			&action.Request{
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
