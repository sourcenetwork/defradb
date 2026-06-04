// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDebugExplainRequestWithDescendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						GROUP (order: {age: DESC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithAscendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						GROUP (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(groupBy: [name]) {
						name
						GROUP (
							groupBy: [verified],
							order: {verified: ASC}
						){
							verified
							GROUP (order: {age: DESC}) {
								age
							}
						}
					}
				}`,

				ExpectedPatterns: groupPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
