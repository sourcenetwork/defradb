// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleWithSizeConstraint_DoesNotErrorOnSizeViolation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with size constraint",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						pointsList: [Int!]
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
						pointsList
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
						pointsList: [Int!] @constraints(size: 2)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Alice",
					"pointsList": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							name
							pointsList
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "Alice",
							// notice the size constraint is not enforced on views
							"pointsList": []int64{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
