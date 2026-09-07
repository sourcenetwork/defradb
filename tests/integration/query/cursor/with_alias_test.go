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

func TestCursorWithAlias_UsesInnerAliasInResponse(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{Doc: `{"name": "Alice", "age": 25}`},
			&action.AddDoc{Doc: `{"name": "Bob", "age": 30}`},
			&action.AddDoc{Doc: `{"name": "Carol", "age": 35}`},

			&action.Request{
				Request: `query {
					_cursor {
						users: User(first: 2, order: {age: ASC}) {
							name
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
						"users": []map[string]any{
							{"name": "Alice", "age": int64(25)},
							{"name": "Bob", "age": int64(30)},
						},
						"_pageInfo": map[string]any{
							"hasNext": true,
							"hasPrev": false,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
