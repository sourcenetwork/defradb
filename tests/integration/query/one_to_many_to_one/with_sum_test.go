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

package one_to_many_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithSumOnInlineAndSumOnOneToManyField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			// Authors
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"favouritePageNumbers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Has written 1 Book
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			// Books
			&action.AddDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				// Has no Publisher.
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			// Publishers
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of The Rooster Bar",
					"address":    "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"_bookID":    testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Theif Lord",
					"address":    "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"_bookID":    testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						ThisMakesNoSenseToSumButHey: SUM(favouritePageNumbers: {})
						TotalRating: SUM(book: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":                        "John Grisham",
							"ThisMakesNoSenseToSumButHey": int64(-1 + 2 + -1 + 1 + 0),
							"TotalRating":                 float64(4.8 + 4.2),
						},
						{
							"name":                        "Cornelia Funke",
							"ThisMakesNoSenseToSumButHey": int64(0),
							"TotalRating":                 float64(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
