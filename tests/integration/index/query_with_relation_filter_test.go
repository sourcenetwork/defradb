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

func TestQueryWithIndex_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String 
					age: Int
					devices: [Device] 
				} 

				type Device {
					model: String @index
					owner: User
				} 
			`),
			sendRequestAndExplain(`
				User(filter: {
					devices: {model: {_eq: "iPhone 10"}} 
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
				},
				NewExplainAsserter().WithDocFetches(3).WithFieldFetches(6).WithIndexFetches(3),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
