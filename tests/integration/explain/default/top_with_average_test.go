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

var topLevelAveragePattern = dataMap{
	"explain": dataMap{
		"topLevelNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"scanNode": dataMap{},
					},
				},
			},
			{
				"sumNode": dataMap{},
			},
			{
				"countNode": dataMap{},
			},
			{
				"averageNode": dataMap{},
			},
		},
	},
}

func TestDefaultExplainTopLevelAverageRequest(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) top-level average request with filter.",

		Request: `query @explain {
			_avg(
				Author: {
					field: age
				}
			)
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": false,
					"age": 28
				}`,
				`{
					"name": "Bob",
					"verified": true,
					"age": 30
				}`,
			},
		},

		ExpectedPatterns: []dataMap{topLevelAveragePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter": dataMap{
						"age": dataMap{
							"_ne": nil,
						},
					},
					"spans": []dataMap{
						{
							"end":   "/4",
							"start": "/3",
						},
					},
				},
			},
			{
				TargetNodeName:    "sumNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"childFieldName": "age",
							"fieldName":      "Author",
							"filter": dataMap{
								"age": dataMap{
									"_ne": nil,
								},
							},
						},
					},
				},
			},
			{
				TargetNodeName:    "countNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"fieldName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_ne": nil,
								},
							},
						},
					},
				},
			},
			{
				TargetNodeName:     "averageNode",
				IncludeChildNodes:  true,      // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{}, // no attributes
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainTopLevelAverageRequestWithFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) top-level average request with filter.",

		Request: `query @explain {
			_avg(
				Author: {
					field: age,
					filter: {
						age: {
							_gt: 26
						}
					}
				}
			)
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": false,
					"age": 21
				}`,
				`{
					"name": "Bob",
					"verified": false,
					"age": 30
				}`,
				`{
					"name": "Alice",
					"verified": false,
					"age": 32
				}`,
			},
		},

		ExpectedPatterns: []dataMap{topLevelAveragePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter": dataMap{
						"age": dataMap{
							"_gt": int32(26),
							"_ne": nil,
						},
					},
					"spans": []dataMap{
						{
							"end":   "/4",
							"start": "/3",
						},
					},
				},
			},
			{
				TargetNodeName:    "sumNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"childFieldName": "age",
							"fieldName":      "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(26),
									"_ne": nil,
								},
							},
						},
					},
				},
			},
			{
				TargetNodeName:    "countNode",
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"fieldName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(26),
									"_ne": nil,
								},
							},
						},
					},
				},
			},
			{
				TargetNodeName:     "averageNode",
				IncludeChildNodes:  true,      // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{}, // no attributes
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}
