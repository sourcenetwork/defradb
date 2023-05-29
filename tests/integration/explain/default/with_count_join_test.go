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

var countTypeIndexJoinPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"countNode": dataMap{
				"selectNode": dataMap{
					"typeIndexJoin": normalTypeJoinPattern,
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithCountOnOneToManyJoinedField(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with count on a one-to-many joined field.",

		Request: `query @explain {
			Author {
				name
				numberOfBooks: _count(books: {})
			}
		}`,

		Docs: map[int][]string{
			//articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			//books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		ExpectedPatterns: []dataMap{countTypeIndexJoinPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
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
					"rootName":    "author",
					"subTypeName": "books",
				},
			},
			{
				TargetNodeName:    "scanNode", // inside of root
				OccurancesToSkip:  0,
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
			{
				TargetNodeName:    "scanNode", // inside of subType (related type)
				OccurancesToSkip:  1,
				IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"filter":         nil,
					"collectionID":   "2",
					"collectionName": "Book",
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

func TestDefaultExplainRequestWithCountOnOneToManyJoinedFieldWithManySources(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with count on a one-to-many joined field with many sources.",

		Request: `query @explain {
			Author {
				name
				numberOfBooks: _count(
					books: {}
					articles: {}
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

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
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
