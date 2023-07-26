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

func TestMultipleOrderByWithDepthGreaterThanOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple orderby with depth greater than 1.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
			Book (order: {rating: ASC, publisher: {yearOpened: DESC}}) {
				name
				rating
				publisher{
					name
					yearOpened
				}
			}
		}`,
				Results: []map[string]any{
					{
						"name":   "Sooley",
						"rating": 3.2,
						"publisher": map[string]any{
							"name":       "Only Publisher of Sooley",
							"yearOpened": uint64(1999),
						},
					},
					{
						"name":   "The Rooster Bar",
						"rating": 4.0,
						"publisher": map[string]any{
							"name":       "Only Publisher of The Rooster Bar",
							"yearOpened": uint64(2022),
						},
					},
					{
						"name":      "The Associate",
						"rating":    4.2,
						"publisher": nil,
					},
					{
						"name":   "A Time for Mercy",
						"rating": 4.5,
						"publisher": map[string]any{
							"name":       "Only Publisher of A Time for Mercy",
							"yearOpened": uint64(2013),
						},
					},
					{
						"name":   "Theif Lord",
						"rating": 4.8,
						"publisher": map[string]any{
							"name":       "Only Publisher of Theif Lord",
							"yearOpened": uint64(2020),
						},
					},
					{
						"name":   "Painted House",
						"rating": 4.9,
						"publisher": map[string]any{
							"name":       "Only Publisher of Painted House",
							"yearOpened": uint64(1995),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMultipleOrderByWithDepthGreaterThanOneOrderSwitched(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple orderby with depth greater than 1, order switched.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
			Book (order: {publisher: {yearOpened: DESC}, rating: ASC}) {
				name
				rating
				publisher{
					name
					yearOpened
				}
			}
		}`,
				Results: []map[string]any{
					{
						"name":   "The Rooster Bar",
						"rating": 4.0,
						"publisher": map[string]any{
							"name":       "Only Publisher of The Rooster Bar",
							"yearOpened": uint64(2022),
						},
					},
					{
						"name":   "Theif Lord",
						"rating": 4.8,
						"publisher": map[string]any{
							"name":       "Only Publisher of Theif Lord",
							"yearOpened": uint64(2020),
						},
					},
					{
						"name":   "A Time for Mercy",
						"rating": 4.5,
						"publisher": map[string]any{
							"name":       "Only Publisher of A Time for Mercy",
							"yearOpened": uint64(2013),
						},
					},
					{
						"name":   "Sooley",
						"rating": 3.2,
						"publisher": map[string]any{
							"name":       "Only Publisher of Sooley",
							"yearOpened": uint64(1999),
						},
					},
					{
						"name":   "Painted House",
						"rating": 4.9,
						"publisher": map[string]any{
							"name":       "Only Publisher of Painted House",
							"yearOpened": uint64(1995),
						},
					},
					{
						"name":      "The Associate",
						"rating":    4.2,
						"publisher": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
