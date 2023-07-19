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
						_key
						name
						age
					}
				}`,

				ExpectedFullGraph: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"_keys":  nil,
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
	}

	explainUtils.ExecuteTestCase(t, test)
}
