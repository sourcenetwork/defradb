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
		Description: "Simple create many mutation",
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
							"_docID": "bae-656366a3-0538-5c4a-a950-b8d48efdb9d8",
							"name":   "John",
							"age":    int64(27),
						},
						{
							"_docID": "bae-b082f68d-ad3e-5b24-a78c-e9ec5c7f8bb6",
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
