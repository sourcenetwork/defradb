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

var averageTypeIndexJoinPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"averageNode": dataMap{
						"countNode": dataMap{
							"sumNode": dataMap{
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
}

func TestDefaultExplainRequestWithAverageOnJoinedField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_avg(books: {field: pages})
					}
				}`,

				ExpectedPatterns: averageTypeIndexJoinPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
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
									"fieldName": "books",
									"filter": dataMap{
										"pages": dataMap{
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
									"childFieldName": "pages",
									"fieldName":      "books",
									"filter": dataMap{
										"pages": dataMap{
											"_ne": nil,
										},
									},
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
						TargetNodeName:    "scanNode", // inside of root type
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter":         nil,
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
							"collectionID":   "bafyreifnc6yphaqxf7x7fa3phxrsuvzqvnnjz4q7fuirhty4cnrxubp6eq",
							"collectionName": "Book",
							"filter": dataMap{
								"pages": dataMap{
									"_ne": nil,
								},
							},
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

func TestDefaultExplainRequestWithAverageOnMultipleJoinedFieldsWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_avg(
							books: {field: pages},
							articles: {field: pages, filter: {pages: {_gt: 3}}}
						)
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"averageNode": dataMap{
										"countNode": dataMap{
											"sumNode": dataMap{
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

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
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
									"fieldName": "books",
									"filter": dataMap{
										"pages": dataMap{
											"_ne": nil,
										},
									},
								},
								{
									"fieldName": "articles",
									"filter": dataMap{
										"pages": dataMap{
											"_gt": int32(3),
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
									"childFieldName": "pages",
									"fieldName":      "books",
									"filter": dataMap{
										"pages": dataMap{
											"_ne": nil,
										},
									},
								},
								{
									"childFieldName": "pages",
									"fieldName":      "articles",
									"filter": dataMap{
										"pages": dataMap{
											"_gt": int32(3),
											"_ne": nil,
										},
									},
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
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
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
							"collectionID":   "bafyreifnc6yphaqxf7x7fa3phxrsuvzqvnnjz4q7fuirhty4cnrxubp6eq",
							"collectionName": "Book",
							"filter": dataMap{
								"pages": dataMap{
									"_ne": nil,
								},
							},
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
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
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
							"collectionID":   "bafyreidpjzf5kytlap2nilnnemywjmwt74hr56e72llri2z7c6w7un7sje",
							"collectionName": "Article",
							"filter": dataMap{
								"pages": dataMap{
									"_gt": int32(3),
									"_ne": nil,
								},
							},
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
// _avg reuses the rendered join as they have matching filters (average adds a ne nil filter).
func TestDefaultExplainRequestOneToManyWithAverageAndChildNeNilFilterSharesJoinField(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						_avg(books: {field: rating})
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
									"averageNode": dataMap{
										"countNode": dataMap{
											"sumNode": dataMap{
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
