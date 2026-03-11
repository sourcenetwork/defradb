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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexList_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(name: "age_index", includes: [{field: "age"}]) {
						name: String @index(name: "name_index")
						age: Int
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "name_index",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
					{
						Name: "age_index",
						ID:   2,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexList_GetIndexesForACollection_ReturnCollectionSpecificList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index
					}

					type Address {
						street: String
						postalCode: String @index
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_age_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "age",
								Descending: false,
							},
						},
						Unique: false,
					},
				},
			},
			&action.ListIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "Address_postalCode_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "postalCode",
								Descending: false,
							},
						},
						Unique: false,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
