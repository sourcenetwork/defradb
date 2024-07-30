// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDefaultExplainOnWrongFieldDirective_BadUsage(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) a request by providing the directive on wrong location (field).",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query {
					Author @explain {
						name
						age
					}
				}`,

				ExpectedError: "Directive \"explain\" may not be used on FIELD.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithFullBasicGraph(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) a basic request.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						age
					}
				}`,

				ExpectedFullGraph: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"docIDs": nil,
										"filter": nil,
										"scanNode": dataMap{
											"filter":         nil,
											"collectionID":   "3",
											"collectionName": "Author",
											"spans": []dataMap{
												{
													"start": "/3",
													"end":   "/4",
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
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainWithAlias(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) a basic request with alias, no filter",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						username: name
						age: age
					}
				}`,

				ExpectedPatterns: basicPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
