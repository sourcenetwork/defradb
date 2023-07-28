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

func TestQueryWithIndex_IfIndexFilterWithRegular_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there is only one indexed field in the query, it should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {
							name: {_in: ["Fred", "Islam", "Addo"]}, 
							age:  {_gt: 40}
						}) {
							name
						}
					}`,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
