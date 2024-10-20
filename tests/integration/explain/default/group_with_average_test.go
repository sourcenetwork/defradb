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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var groupAveragePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"averageNode": dataMap{
						"countNode": dataMap{
							"sumNode": dataMap{
								"groupNode": dataMap{
									"selectNode": dataMap{
										"scanNode": dataMap{},
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

func TestDefaultExplainRequestWithGroupByWithAverageOnAnInnerField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with group-by with average on inner field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						_avg(_group: {field: age})
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"docID":          nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
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
						IncludeChildNodes:  false,
						ExpectedAttributes: dataMap{}, // no attributes
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "_group",
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
						TargetNodeName:    "sumNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "age",
									"fieldName":      "_group",
									"filter": dataMap{
										"age": dataMap{
											"_ne": nil,
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

func TestDefaultExplainRequestWithAverageInsideTheInnerGroupOnAField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with group-by with average of the inner _group on a field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						_avg(_group: {field: _avg})
						_group(groupBy: [verified]) {
							verified
							_avg(_group: {field: age})
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"groupBy":        []string{"verified", "name"},
									"docID":          nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:     "averageNode",
						IncludeChildNodes:  false,
						ExpectedAttributes: dataMap{}, // no attributes
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "_group",
									"filter":    nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "sumNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "_avg",
									"fieldName":      "_group",
									"filter":         nil,
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

func TestDefaultExplainRequestWithAverageInsideTheInnerGroupOnAFieldAndNestedGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with group-by with average of the inner _group on a field and nested group-by.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						_avg(_group: {field: _avg})
						_group(groupBy: [verified]) {
							verified
								_avg(_group: {field: age})
								_group (groupBy: [age]){
									age
								}
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"groupBy":        []string{"verified", "name"},
									"docID":          nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:     "averageNode",
						IncludeChildNodes:  false,
						ExpectedAttributes: dataMap{}, // no attributes
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "_group",
									"filter":    nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "sumNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "_avg",
									"fieldName":      "_group",
									"filter":         nil,
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

func TestDefaultExplainRequestWithAverageInsideTheInnerGroupAndNestedGroupByWithAverage(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with average inside the inner _group and nested groupBy with average.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						_avg(_group: {field: _avg})
						_group(groupBy: [verified]) {
							verified
								_avg(_group: {field: age})
								_group (groupBy: [age]){
									age
									_avg(_group: {field: age})
								}
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"groupBy":        []string{"verified", "name"},
									"docID":          nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter":         nil,
								},
							},
						},
					},
					{
						TargetNodeName:     "averageNode",
						IncludeChildNodes:  false,
						ExpectedAttributes: dataMap{}, // no attributes
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "_group",
									"filter":    nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "sumNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "_avg",
									"fieldName":      "_group",
									"filter":         nil,
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
