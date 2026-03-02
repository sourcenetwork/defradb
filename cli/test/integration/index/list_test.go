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

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

func TestIndexList_WithEmptyCollection_ShouldReturnEmptyList(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.ListIndexes{
				Collection:      "User",
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	test.Execute(t)
}

func TestIndexList_WithSingleCollection_ShouldReturnAllCollectionIndexes(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByAge",
				Fields:     []string{"age:DESC"},
			},
			&action.ListIndexes{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByAge",
						Fields: []client.IndexedFieldDescription{
							{Name: "age", Descending: true},
						},
						Unique: false,
					},
				},
			},
		},
	}

	test.Execute(t)
}

func TestIndexList_WithoutCollectionFlag_ShouldReturnAllIndexes(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type User {
						name: String
						age: Int
					}
					
					type Product {
						title: String
						price: Float
					}
				`,
			},
			// Make indexes for User collection
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
			},
			&action.NewIndex{
				Collection: "User",
				Name:       "UsersByAge",
				Fields:     []string{"age"},
			},
			// Make indexes for Product collection
			&action.NewIndex{
				Collection: "Product",
				Name:       "ProductsByTitle",
				Fields:     []string{"title"},
			},
			&action.NewIndex{
				Collection: "Product",
				Name:       "ProductsByPrice",
				Fields:     []string{"price:DESC"},
				Unique:     true,
			},
			// List all indexes
			&action.ListIndexes{
				ExpectedAllIndexes: map[client.CollectionName][]client.IndexDescription{
					"User": {
						{
							Name: "UsersByName",
							Fields: []client.IndexedFieldDescription{
								{Name: "name", Descending: false},
							},
							Unique: false,
						},
						{
							Name: "UsersByAge",
							Fields: []client.IndexedFieldDescription{
								{Name: "age", Descending: false},
							},
							Unique: false,
						},
					},
					"Product": {
						{
							Name: "ProductsByTitle",
							Fields: []client.IndexedFieldDescription{
								{Name: "title", Descending: false},
							},
							Unique: false,
						},
						{
							Name: "ProductsByPrice",
							Fields: []client.IndexedFieldDescription{
								{Name: "price", Descending: true},
							},
							Unique: true,
						},
					},
				},
			},
		},
	}

	test.Execute(t)
}

func TestIndexList_WithEmptyDatabase_ShouldReturnEmptyMap(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			// List all indexes when no collections exist
			&action.ListIndexes{
				ExpectedAllIndexes: map[client.CollectionName][]client.IndexDescription{},
			},
		},
	}

	test.Execute(t)
}

func TestIndexList_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.ListIndexes{
				Collection:  "NonExistentCollection",
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}
