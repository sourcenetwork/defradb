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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var countTypeIndexJoinPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"countNode": dataMap{
						"selectNode": dataMap{
							"typeIndexJoin": normalTypeJoinPattern,
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithCountOnOneToManyJoinedField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with count on a one-to-many joined field.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						numberOfBooks: _count(books: {})
					}
				}`,

				ExpectedPatterns: countTypeIndexJoinPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"filter":    nil,
									"fieldName": "books",
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "books",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of root
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"filter":         nil,
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of subType (related type)
						OccurancesToSkip:  1,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"filter":         nil,
							"collectionID":   "bafkreictgwjdhpmwh6a7uu65oggzcgolwiq5vb53x2ehhijh2yqzgwzjwu",
							"collectionName": "Book",
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

func TestDefaultExplainRequestWithCountOnOneToManyJoinedFieldWithManySources(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with count on a one-to-many joined field with many sources.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						numberOfBooks: _count(
							books: {}
							articles: {}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
										"selectNode": dataMap{
											"parallelNode": []dataMap{
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"filter":    nil,
									"fieldName": "books",
								},

								{
									"filter":    nil,
									"fieldName": "articles",
								},
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "books",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 1st root type
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 1st subType (related type)
						OccurancesToSkip:  1,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreictgwjdhpmwh6a7uu65oggzcgolwiq5vb53x2ehhijh2yqzgwzjwu",
							"collectionName": "Book",
							"filter":         nil,
							"prefixes": []string{
								"/2",
							},
						},
					},
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "articles",
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 2nd root type (AKA: subType's root)
						OccurancesToSkip:  2,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						TargetNodeName:    "scanNode", // inside of 2nd subType (AKA: subType's subtype)
						OccurancesToSkip:  3,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreihlfvtpy72o354ig4qqvyfeh2gelyijemw2brtfyq6cwuglaro5ba",
							"collectionName": "Article",
							"filter":         nil,
							"prefixes": []string{
								"/1",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

// This test asserts that only a single index join is used (not parallelNode) because the
// _count reuses the rendered join as they have matching filters.
func TestDefaultExplainRequestOneToManyWithCountWithFilterAndChildFilterSharesJoinField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) 1-to-M relation request from many side with count filter shared.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_count(books: {filter: {rating: {_ne: null}}})
						books(filter: {rating: {_ne: null}}){
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
										"selectNode": dataMap{
											"typeIndexJoin": normalTypeJoinPattern,
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

// This test asserts that two joins are used (with parallelNode) because _count cannot
// reuse the rendered join as they dont have matching filters.
func TestDefaultExplainRequestOneToManyWithCountAndChildFilterDoesNotShareJoinField(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) 1-to-M relation request from many side with count filter not shared.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_count(books: {})
						books(filter: {rating: {_ne: null}}){
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
										"selectNode": dataMap{
											"parallelNode": []dataMap{
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
											},
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
