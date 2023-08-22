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
					name: String @index
					age: Int
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {
					name: {_in: ["Fred", "Islam", "Addo"]}, 
					age:  {_gt: 40}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
				},
				NewExplainAsserter().WithDocFetches(3).WithFieldFetches(6),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfMultipleIndexFiltersWithRegular_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there is only one indexed field in the query, it should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int @index
					email: String 
				} 
			`),
			sendRequestAndExplain(`
				users(filter: {
					name: {_like: "%a%"}, 
					age:  {_gt: 30},
					email: {_like: "%m@gmail.com"}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
				},
				NewExplainAsserter().WithDocFetches(5).WithFieldFetches(15),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
