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

func TestDebugExplainRequestWithRelatedAndRegularFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with related and regular filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						filter: {
							name: {_eq: "John Grisham"},
							books: {name: {_eq: "Painted House"}}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"typeIndexJoin": dataMap{
										"typeJoinMany": normalTypeJoinPattern,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainRequestWithManyRelatedFilters(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with many related filters.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain(type: debug) {
					Author(
						filter: {
							name: {_eq: "Cornelia Funke"},
							articles: {name: {_eq: "To my dear readers"}},
							books: {name: {_eq: "Theif Lord"}}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"parallelNode": []dataMap{
										{
											"typeIndexJoin": dataMap{
												"typeJoinMany": debugTypeJoinPattern,
											},
										},
										{
											"typeIndexJoin": dataMap{
												"typeJoinMany": debugTypeJoinPattern,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
