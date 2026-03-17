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

func TestMultipleOrderByWithDepthGreaterThanOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			addDocsWith6BooksAnd5Publishers(),
			&action.Request{
				Request: `query {
			Book (order: [{rating: ASC}, {publisher: {yearOpened: DESC}}]) {
				name
				rating
				publisher{
					name
					yearOpened
				}
			}
		}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Sooley",
							"rating": 3.2,
							"publisher": map[string]any{
								"name":       "Only Publisher of Sooley",
								"yearOpened": int64(1999),
							},
						},
						{
							"name":   "The Rooster Bar",
							"rating": 4.0,
							"publisher": map[string]any{
								"name":       "Only Publisher of The Rooster Bar",
								"yearOpened": int64(2022),
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
								"yearOpened": int64(2013),
							},
						},
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"publisher": map[string]any{
								"name":       "Only Publisher of Theif Lord",
								"yearOpened": int64(2020),
							},
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
							"publisher": map[string]any{
								"name":       "Only Publisher of Painted House",
								"yearOpened": int64(1995),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The @exhaustive directive is needed because the primary order is on a relation field
// (publisher.yearOpened). When the secondary-index multiplier adds an index on yearOpened,
// the planner inverts the join and orphan parents (books without publishers) are excluded.
// Without @exhaustive, this test would need MultiplierExcludes for secondary-index.
func TestMultipleOrderByWithDepthGreaterThanOneOrderSwitched(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			addDocsWith6BooksAnd5Publishers(),
			&action.Request{
				Request: `query @exhaustive {
					Book (order: [{publisher: {yearOpened: DESC}}, {rating: ASC}]) {
						name
						rating
						publisher{
							name
							yearOpened
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "The Rooster Bar",
							"rating": 4.0,
							"publisher": map[string]any{
								"name":       "Only Publisher of The Rooster Bar",
								"yearOpened": int64(2022),
							},
						},
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"publisher": map[string]any{
								"name":       "Only Publisher of Theif Lord",
								"yearOpened": int64(2020),
							},
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
							"publisher": map[string]any{
								"name":       "Only Publisher of A Time for Mercy",
								"yearOpened": int64(2013),
							},
						},
						{
							"name":   "Sooley",
							"rating": 3.2,
							"publisher": map[string]any{
								"name":       "Only Publisher of Sooley",
								"yearOpened": int64(1999),
							},
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
							"publisher": map[string]any{
								"name":       "Only Publisher of Painted House",
								"yearOpened": int64(1995),
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
