// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithSumOnInlineAndSumOnOneToManyField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Sum of integer array, and sum of one-to-many field.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			// Authors
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-3c4217d2-f879-50b1-b375-acf42b764e5b, Has written 5 books
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"favouritePageNumbers": [-1, 2, -1, 1, 0]
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 Book
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			// Books
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 Publisher
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-afdd1769-b056-5bb1-b743-116a347b4b87", Has 1 Publisher
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-3c4217d2-f879-50b1-b375-acf42b764e5b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-fbba03cf-c77c-5850-a6a4-0d9992d489e1", Has no Publisher.
				Doc: `{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-3c4217d2-f879-50b1-b375-acf42b764e5b"
				}`,
			},
			// Publishers
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-afdd1769-b056-5bb1-b743-116a347b4b87"
			    }`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						ThisMakesNoSenseToSumButHey: _sum(favouritePageNumbers: {})
						TotalRating: _sum(book: {field: rating})
					}
				}`,
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTEMP(t, test)
}
