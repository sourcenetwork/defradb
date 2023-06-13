// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDebugExplainRequestWithDescendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order (descending) on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						_group (order: {age: DESC}){
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAscendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order (ascending) on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						_group (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with order on nested parent groupBy and on nested parent's inner _group.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						_group (
							groupBy: [verified],
							order: {verified: ASC}
						){
							verified
							_group (order: {age: DESC}) {
								age
							}
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupPattern},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
