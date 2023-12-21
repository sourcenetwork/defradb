// Copyright 2023 Democratized Data Foundation
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

func TestView_SimpleWithAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with alias",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						fullname: name
					}
				`,
				SDL: `
					type UserView {
						fullname: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							fullname
						}
					}
				`,
				Results: []map[string]any{
					{
						"fullname": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
