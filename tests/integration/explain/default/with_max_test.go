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

	"github.com/sourcenetwork/defradb/tests/action"
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Book {
						name
						MaxChapterPages: MAX(chapterPages: {})
					}
				}`,

				ExpectedPatterns: maxPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
							"collectionID":   "bafyreifnc6yphaqxf7x7fa3phxrsuvzqvnnjz4q7fuirhty4cnrxubp6eq",
							"collectionName": "Book",
							"filter":         nil,
							"prefixes": []string{
								"/2",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
