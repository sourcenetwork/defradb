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

func TestDefaultExplainRequestWithDockeysOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys on inner _group.",

		Request: `query @explain {
			Author(
				groupBy: [age]
			) {
				age
				_group(dockeys: ["bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"]) {
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

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"docKeys":        []string{"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"},
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
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithDockeyOnParentGroupBy(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with a dockey on parent groupBy.",

		Request: `query @explain {
			Author(
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

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
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
									"collectionName": "Author",
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

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithDockeysAndFilterOnParentGroupBy(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys and filter on parent groupBy.",

		Request: `query @explain {
			Author(
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

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
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
									"collectionName": "Author",
									"filter": dataMap{
										"age": dataMap{
											"_eq": int32(20),
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

	runExplainTest(t, test)
}
