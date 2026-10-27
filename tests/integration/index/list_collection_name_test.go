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

// Listing a single collection's indexes returns only an opaque CollectionID unless each
// result also names its collection, so the caller cannot render the owner without a
// second lookup. See https://github.com/sourcenetwork/defradb/issues/5055.
func TestIndexList_WithCollectionScope_ResultsNameTheirCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "UsersByName",
				Fields:       []client.IndexedFieldDescription{{Name: "name"}},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "UsersByName",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
				ExpectedCollectionName: "User",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Each collection's results must name that collection, not merely the first one, so a
// result lifted out of a multi-collection listing stays self-describing.
func TestIndexList_WithMultipleCollections_EachResultNamesItsOwnCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}

					type Product {
						title: String
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "UsersByName",
				Fields:       []client.IndexedFieldDescription{{Name: "name"}},
			},
			&action.NewIndex{
				CollectionID: 1,
				IndexName:    "ProductsByTitle",
				Fields:       []client.IndexedFieldDescription{{Name: "title"}},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "UsersByName",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
				ExpectedCollectionName: "User",
			},
			&action.ListIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "ProductsByTitle",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "title"}},
					},
				},
				ExpectedCollectionName: "Product",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
