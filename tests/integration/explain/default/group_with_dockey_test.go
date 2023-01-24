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
)

func TestExplainQueryWithDockeyOnParentGroupBy(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with a dockey on parent groupBy.",

		Query: `query @explain {
			author(
				groupBy: [age],
				dockey: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
			) {
				age
				_group {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// dockey: "bae-21a6ad4a-1cd8-5613-807c-a90c7c12f880"
				`{
					"name": "John Grisham",
					"age": 12
				}`,

				// dockey: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
				`{
					"name": "Cornelia Funke",
					"age": 20
				}`,

				// dockey: "bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
				`{
					"name": "John's Twin",
					"age": 65
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "author",
									"docKeys":        nil,
									"filter":         nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
								},
							},
							"groupByFields": []string{"age"},
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
											"end":   "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2255",
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

	executeTestCase(t, test)
}

func TestExplainQuerySimpleWithDockeysAndFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with a dockeys and filter on parent groupBy.",

		Query: `query @explain {
			author(
				groupBy: [age],
				filter: {age: {_eq: 20}},
				dockeys: [
					"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
					"bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
				]
			) {
				age
				_group {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				// dockey: "bae-21a6ad4a-1cd8-5613-807c-a90c7c12f880"
				`{
                     "name": "John Grisham",
                     "age": 12
                 }`,

				// dockey: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
				`{
                     "name": "Cornelia Funke",
                     "age": 20
                 }`,

				// dockey: "bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
				`{
                     "name": "John's Twin",
                     "age": 65
                 }`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "author",
									"docKeys":        nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter":         nil,
								},
							},
							"groupByFields": []string{"age"},
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter": dataMap{
										"age": dataMap{
											"_eq": int(20),
										},
									},
									"spans": []dataMap{
										{
											"start": "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
											"end":   "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2255",
										},
										{
											"start": "/3/bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed",
											"end":   "/3/bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeee",
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

	executeTestCase(t, test)
}
