// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
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
			testUtils.Request{
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
