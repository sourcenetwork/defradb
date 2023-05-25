// Copyright 2023 Democratized Data Foundation
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

func TestDefaultExplainRequestWithDockeysOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys on inner _group.",

		Request: `query @explain {
			Author(
				groupBy: [age]
			) {
				age
				_group(dockeys: ["bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"]) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// dockey: "bae-21a6ad4a-1cd8-5613-807c-a90c7c12f880"
				`{
					"name": "John Grisham",
					"age": 12
				}`,

				// dockey: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
				`{
					"name": "Cornelia Funke",
					"age": 20
				}`,

				// dockey: "bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
				`{
					"name": "John's Twin",
					"age": 65
				}`,
			},
		},

		ExpectedPatterns: []dataMap{groupPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "groupNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"groupByFields": []string{"age"},
					"childSelects": []dataMap{
						{
							"collectionName": "Author",
							"docKeys":        []string{"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"},
							"filter":         nil,
							"groupBy":        nil,
							"limit":          nil,
							"orderBy":        nil,
						},
					},
				},
			},
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/3",
							"end":   "/4",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}
