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

func TestDebugExplainRequestWithStringEqualFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {name: {_eq: "Lone"}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithIntegerEqualFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {age: {_eq: 26}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithGreaterThanFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {age: {_gt: 20}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLogicalCompoundAndFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {_and: [{age: {_gt: 20}}, {age: {_lt: 50}}]}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithLogicalCompoundOrFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {_or: [{age: {_eq: 55}}, {age: {_eq: 19}}]}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithMatchInsideList(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(filter: {age: {_in: [19, 40, 55]}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
