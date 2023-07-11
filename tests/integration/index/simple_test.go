// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexWithExplain(t *testing.T) {
	test := testUtils.TestCase{
		Description: "",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String @index
					} 
				`,
			},
			createUserDocs(),
			testUtils.Request{
				Request: `
					query @explain(type: execute) {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Asserter: newExplainAsserter(2, 2, 1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						name: String @index
					} 
				`,
			},
			createUserDocs(),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Islam",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}