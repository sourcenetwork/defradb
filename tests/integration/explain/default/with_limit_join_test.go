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

var limitTypeJoinPattern = dataMap{
	"root": dataMap{
		"scanNode": dataMap{},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"limitNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithOnlyLimitOnRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with only limit on related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						articles(limit: 1) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": limitTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(1),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithOnlyOffsetOnRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with only offset on related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						articles(offset: 2) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": limitTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  nil,
							"offset": uint64(2),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithBothLimitAndOffsetOnRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with both limit and offset on related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						articles(limit: 2, offset: 2) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": limitTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(2),
							"offset": uint64(2),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLimitOnRelatedChildAndBothLimitAndOffsetOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit on related child & both limit + offset on parent.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(limit: 3, offset: 1) {
						name
						articles(limit: 2) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"limitNode": dataMap{
										"selectNode": dataMap{
											"typeIndexJoin": limitTypeJoinPattern,
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
						},
					},
					{
						TargetNodeName:    "limitNode",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(2),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
