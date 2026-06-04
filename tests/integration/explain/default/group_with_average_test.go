// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_explain_default

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: age})
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
											"_neq": nil,
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
									"fieldName": "GROUP",
									"filter": dataMap{
										"age": dataMap{
											"_neq": nil,
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
									"fieldName":      "GROUP",
									"filter": dataMap{
										"age": dataMap{
											"_neq": nil,
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
							AVG(GROUP: {field: age})
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
									"fieldName": "GROUP",
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
									"childFieldName": "AVG",
									"fieldName":      "GROUP",
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
								AVG(GROUP: {field: age})
								GROUP (groupBy: [age]){
									age
								}
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
									"fieldName": "GROUP",
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
									"childFieldName": "AVG",
									"fieldName":      "GROUP",
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author (groupBy: [name]) {
						name
						AVG(GROUP: {field: AVG})
						GROUP(groupBy: [verified]) {
							verified
								AVG(GROUP: {field: age})
								GROUP (groupBy: [age]){
									age
									AVG(GROUP: {field: age})
								}
						}
					}
				}`,

				ExpectedPatterns: groupAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
									"fieldName": "GROUP",
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
									"childFieldName": "AVG",
									"fieldName":      "GROUP",
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
