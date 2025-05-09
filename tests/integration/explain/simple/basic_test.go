// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

type dataMap = map[string]any

func TestSimpleExplainRequest(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (simple) a basic request, assert full graph.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{
				Request: `query @explain(type: simple) {
					Author {
						_docID
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
										"docID":  nil,
										"filter": nil,
										"scanNode": dataMap{
											"filter":         nil,
											"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
											"collectionName": "Author",
											"prefixes": []string{
												"/3",
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
