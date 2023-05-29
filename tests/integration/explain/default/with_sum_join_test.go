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

var sumTypeIndexJoinPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"sumNode": dataMap{
				"selectNode": dataMap{
					"typeIndexJoin": normalTypeJoinPattern,
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithSumOnOneToManyJoinedField(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with sum on a one-to-many joined field.",

		Request: `query @explain {
			Author {
				name
				_key
				TotalPages: _sum(
					books: {field: pages}
				)
			}
		}`,

		Docs: map[int][]string{
			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 22
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 101
				}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321
				}`,
			},

			// authors
			2: {
				// _key: "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
		},

		ExpectedPatterns: []dataMap{sumTypeIndexJoinPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "sumNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"fieldName":      "books",
							"childFieldName": "pages",
							"filter":         nil,
						},
					},
				},
			},
			{
				TargetNodeName:    "typeIndexJoin",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"joinType":    "typeJoinMany",
					"rootName":    "author",
					"subTypeName": "books",
				},
			},
			{
				TargetNodeName:    "scanNode", // inside of root
				OccurancesToSkip:  0,
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
			{
				TargetNodeName:    "scanNode", // inside of subType (related type)
				OccurancesToSkip:  1,
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

func TestDefaultExplainRequestWithSumOnOneToManyJoinedFieldWithFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with sum on a one-to-many joined field, with filter.",

		Request: `query @explain {
			Author {
				name
				TotalPages: _sum(
					articles: {
						field: pages,
						filter: {
							name: {
								_eq: "To my dear readers"
							}
						}
					}
				)
			}
		}`,

		Docs: map[int][]string{
			// articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 2
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 11
				}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 31
				}`,
			},

			// authors
			2: {
				// _key: "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
		},

		ExpectedPatterns: []dataMap{sumTypeIndexJoinPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "sumNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"fieldName":      "articles",
							"childFieldName": "pages",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "To my dear readers",
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
					"rootName":    "author",
					"subTypeName": "articles",
				},
			},
			{
				TargetNodeName:    "scanNode", // inside of root
				OccurancesToSkip:  0,
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
			{
				TargetNodeName:    "scanNode", // inside of subType (related type)
				OccurancesToSkip:  1,
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "1",
					"collectionName": "Article",
					"filter": dataMap{
						"name": dataMap{
							"_eq": "To my dear readers",
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
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithSumOnOneToManyJoinedFieldWithManySources(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with sum on a one-to-many joined field with many sources.",

		Request: `query @explain {
			Author {
				name
				TotalPages: _sum(
					books: {field: pages},
					articles: {field: pages}
				)
			}
		}`,

		Docs: map[int][]string{
			// articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 2
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 11
				}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 31
				}`,
			},

			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 22
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 101
				}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321
				}`,
			},

			// authors
			2: {
				// _key: "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
		},

		ExpectedPatterns: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
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

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "sumNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"sources": []dataMap{
						{
							"childFieldName": "pages",
							"fieldName":      "books",
							"filter":         nil,
						},

						{
							"childFieldName": "pages",
							"fieldName":      "articles",
							"filter":         nil,
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
					"rootName":    "author",
					"subTypeName": "books",
				},
			},
			{
				TargetNodeName:    "scanNode", // inside of 1st root type
				OccurancesToSkip:  0,
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
			{
				TargetNodeName:    "scanNode", // inside of 1st subType (related type)
				OccurancesToSkip:  1,
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
			{
				TargetNodeName:    "typeIndexJoin",
				OccurancesToSkip:  1,
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"joinType":    "typeJoinMany",
					"rootName":    "author",
					"subTypeName": "articles",
				},
			},
			{
				TargetNodeName:    "scanNode", // inside of 2nd root type (AKA: subType's root)
				OccurancesToSkip:  2,
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
			{
				TargetNodeName:    "scanNode", // inside of 2nd subType (AKA: subType's subtype)
				OccurancesToSkip:  3,
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "1",
					"collectionName": "Article",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/1",
							"end":   "/2",
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}
