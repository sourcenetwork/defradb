// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"
)

func TestQueryWithIndex_IfIndexFilterWithRegular_ShouldFilter(t *testing.T) {
	req := `query {
		User(filter: {
			name: {_in: ["Fred", "Islam", "Addo"]}, 
			age:  {_gt: 40}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Combination of a filter on regular and of an indexed field",
		Actions: []any{
			testUtils.GenerateDocsFromSchema{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
				CreateSchema:   true,
				PredefinedDocs: immutable.Some(getUserDocs()),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(3).WithFieldFetches(6).WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfMultipleIndexFiltersWithRegular_ShouldFilter(t *testing.T) {
	req := `query {
		User(filter: {
			name: {_like: "%a%"}, 
			age:  {_gt: 30},
			email: {_like: "%m@gmail.com"}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Combination of a filter on regular and of 2 indexed fields",
		Actions: []any{
			testUtils.GenerateDocsFromSchema{
				Schema: `
					type User {
						name: String @index
						age: Int @index
						email: String 
					}`,
				CreateSchema:   true,
				PredefinedDocs: immutable.Some(getUserDocs()),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Islam"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(6).WithFieldFetches(18),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
