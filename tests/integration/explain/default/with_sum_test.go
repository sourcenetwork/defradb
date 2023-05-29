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

		Docs: map[int][]string{
			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 77,
					"chapterPages": [1, 22, 33, 44, 55, 66]
				}`, // sum of chapterPages == 221

				`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 55,
					"chapterPages": [1, 22]
				}`, // sum of chapterPages == 23

				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321,
					"chapterPages": [10, 50, 100, 200, 300]
				}`, // sum of chapterPages == 660
			},
		},

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
