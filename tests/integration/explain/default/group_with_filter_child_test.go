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

func TestDefaultExplainRequestWithFilterOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with filter on the inner _group selection.",

		Request: `query @explain {
			Author (groupBy: [age]) {
				age
				_group(filter: {age: {_gt: 63}}) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
                     "name": "John Grisham",
                     "age": 65
                 }`,

				`{
                     "name": "Cornelia Funke",
                     "age": 62
                 }`,

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
							"docKeys":        nil,
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(63),
								},
							},
							"groupBy": nil,
							"limit":   nil,
							"orderBy": nil,
						},
					},
				},
			},
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithFilterOnParentGroupByAndInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with filter on parent groupBy and on the inner _group selection.",

		Request: `query @explain {
			Author (
				groupBy: [age],
				filter: {age: {_gt: 62}}
			) {
				age
				_group(filter: {age: {_gt: 63}}) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
                     "name": "John Grisham",
                     "age": 65
                 }`,

				`{
                     "name": "Cornelia Funke",
                     "age": 62
                 }`,

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
							"docKeys":        nil,
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(63),
								},
							},
							"groupBy": nil,
							"limit":   nil,
							"orderBy": nil,
						},
					},
				},
			},
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"filter": dataMap{
						"age": dataMap{
							"_gt": int32(62),
						},
					},
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
	}

	runExplainTest(t, test)
}
