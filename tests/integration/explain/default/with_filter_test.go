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

func TestDefaultExplainRequestWithStringEqualFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with string equal (_eq) filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {name: {_eq: "Lone"}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Lone",
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithIntegerEqualFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with integer equal (_eq) filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {age: {_eq: 26}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_eq": int32(26),
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithGreaterThanFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with greater than (_gt) filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {age: {_gt: 20}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(20),
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLogicalCompoundAndFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with logical compound (_and) filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {_and: [{age: {_gt: 20}}, {age: {_lt: 50}}]}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"_and": []any{
									dataMap{
										"age": dataMap{
											"_gt": int32(20),
										},
									},
									dataMap{
										"age": dataMap{
											"_lt": int32(50),
										},
									},
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLogicalCompoundOrFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with logical compound (_or) filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {_or: [{age: {_eq: 55}}, {age: {_eq: 19}}]}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"_or": []any{
									dataMap{
										"age": dataMap{
											"_eq": int32(55),
										},
									},
									dataMap{
										"age": dataMap{
											"_eq": int32(19),
										},
									},
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithMatchInsideList(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request filtering values that match within (_in) a list.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(filter: {age: {_in: [19, 40, 55]}}) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_in": []any{
										int32(19),
										int32(40),
										int32(55),
									},
								},
							},
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
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequest_WithJSONEqualFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (default) request with JSON equal (_eq) filter.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			testUtils.ExplainRequest{
				Request: `query @explain {
					Users(filter: {custom: {_eq: {one: {two: 3}}}}) {
						name
					}
				}`,
				ExpectedPatterns: basicPattern,
				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "1",
							"collectionName": "Users",
							"filter": dataMap{
								"custom": dataMap{
									"_eq": dataMap{
										"one": dataMap{
											"two": int32(3),
										},
									},
								},
							},
							"spans": []dataMap{
								{
									"start": "/1",
									"end":   "/2",
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
