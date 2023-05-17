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

	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDefaultExplainOnWrongFieldDirective_BadUsage(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) a request by providing the directive on wrong location (field).",

		Request: `query {
			Author @explain {
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		ExpectedError: "Directive \"explain\" may not be used on FIELD.",
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithFullBasicGraph(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) a basic request.",

		Request: `query @explain {
			Author {
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainWithAlias(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) a basic request with alias, no filter",

		Request: `query @explain {
			Author {
				username: name
				age: age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		ExpectedPatterns: []dataMap{basicPattern},
	}

	runExplainTest(t, test)
}
