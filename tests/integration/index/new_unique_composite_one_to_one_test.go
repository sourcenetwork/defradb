// Copyright 2025 Democratized Data Foundation
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

func TestOneToOneUniqueIndex_OnPrimarySide_AutoAdded(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User__addressID_ASC",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_addressID"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_UserDefinedUniqueIndexWithName_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						address: Address @primary @index(unique: true, name: "user_address_unique")
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "user_address_unique",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_addressID"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_UserDefinedNonUniqueIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						address: Address @primary @index(unique: false)
					}

					type Address {
						street: String
						user: User
					}`,
				ExpectedError: "one-to-one relation must have a unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_TypeLevelCompositeUniqueIndex_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "_addressID"}, {field: "name"}]) {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User__addressID_ASC",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_addressID"},
							{Name: "name"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_TypeLevelNonUniqueIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: false, includes: [{field: "_addressID"}]) {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
				ExpectedError: "one-to-one relation must have a unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_CompositeIndexRelationNotFirst_AutoIndexStillAdded(t *testing.T) {
	// When user defines a composite index where relation is NOT the first field,
	// the automatic unique index should still be created
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "_addressID"}]) {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name_ASC",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "name"},
							{Name: "_addressID"},
						},
					},
					{
						Name:   "User__addressID_ASC",
						ID:     2,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "_addressID"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_ReferenceSameRelatedDoc_RejectsDuplicateLink(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"street": "123 Main St"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "John",
					"_addressID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "Jane",
					"_addressID": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_MultipleNullRelations_Allowed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						address: Address @primary
					}

					type Address {
						street: String
						user: User
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Jane"
				}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Jane"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToOneUniqueIndex_OneToMany_ShouldNotMakeNewIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						books: [Book]
					}

					type Book {
						title: String
						author: Author
					}`,
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
			&action.ListIndexes{
				CollectionID:    1,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
