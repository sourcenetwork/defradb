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

var sumPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"sumNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithSumOnInlineArrayField_ChildFieldWillBeEmpty(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with sum on an inline array field.",

		Request: `query @explain {
			Book {
				name
				NotSureWhySomeoneWouldSumTheChapterPagesButHereItIs: _sum(chapterPages: {})
			}
		}`,

		ExpectedPatterns: []dataMap{sumPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "sumNode",
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
	}

	explainUtils.RunExplainTest(t, test)
}
