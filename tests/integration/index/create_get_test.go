// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexGet_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Getting indexes should return list of existing indexes",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(name: "age_index", includes: [{field: "age"}]) {
						name: String @index(name: "name_index")
						age: Int
					}
				`,
			},
			testUtils.GetIndexes{
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

func TestIndexGet_GetIndexesForACollection_ReturnCollectionSpecificList(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Getting indexes for a collection should return only the indexes on that collection",
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			testUtils.GetIndexes{
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
			testUtils.GetIndexes{
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
