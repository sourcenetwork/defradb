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

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `[ 
					{
						"name": "John",
						"age": 27
					},
					{
						"name": "Islam",
						"age": 33
					}
				]`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
							"name":   "John",
							"age":    int64(27),
						},
						{
							"_docID": "bae-974c991f-74fb-5841-99a7-7c85a4942fbc",
							"name":   "Islam",
							"age":    int64(33),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
