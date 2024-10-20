// Copyright 2024 Democratized Data Foundation
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

var maxPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"maxNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequest_WithMaxOnInlineArrayField_ChildFieldWillBeEmpty(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with max on an inline array field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Book {
						name
						MaxChapterPages: _max(chapterPages: {})
					}
				}`,

				ExpectedPatterns: maxPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "maxNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName":      "chapterPages",
									"childFieldName": nil,
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "2",
							"collectionName": "Book",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/2",
									"end":   "/3",
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
