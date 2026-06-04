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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Book {
						name
						__typename
						author {
							name
							__typename
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":       "Painted House",
							"__typename": "Book",
							"author": map[string]any{
								"name":       "John Grisham",
								"__typename": "Author",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
