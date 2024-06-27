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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation",
		Actions: []any{
			testUtils.SchemaUpdate{
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
				Results: []map[string]any{
					{
						"_docID": "bae-48339725-ed14-55b1-8e63-3fda5f590725",
						"name":   "Islam",
						"age":    int64(33),
					},
					{
						"_docID": "bae-8c89a573-c287-5d8c-8ba6-c47c814c594d",
						"name":   "John",
						"age":    int64(27),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
