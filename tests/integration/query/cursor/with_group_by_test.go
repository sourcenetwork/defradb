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

package cursor

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCursorWithGroupBy_ReturnsGroupedResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},
			&action.AddDoc{Doc: `{"name": "Dave", "age": 45}`},
			&action.Request{
				Request: `query {
					_cursor {
						User(first: 10, groupBy: [age]) {
							age
						}
						_pageInfo {
							hasNext
							hasPrev
						}
					}
				}`,
				Results: map[string]any{
					"_cursor": map[string]any{
						"User": []map[string]any{
							{"age": int64(25)},
							{"age": int64(35)},
							{"age": int64(45)},
						},
						"_pageInfo": map[string]any{
							"hasNext": false,
							"hasPrev": false,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
