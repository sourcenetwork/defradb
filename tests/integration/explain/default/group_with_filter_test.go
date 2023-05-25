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

func TestDefaultExplainRequestWithFilterOnGroupByParent(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with filter on parent groupBy.",

		Request: `query @explain {
			Author (
				groupBy: [age],
				filter: {age: {_gt: 63}}
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
				`{
                     "name": "John Grisham",
                     "age": 65
                 }`,

				`{
                     "name": "Cornelia Funke",
                     "age": 62
                 }`,

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
							"groupByFields": []string{"age"},
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
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "Author",
									"filter": dataMap{
										"age": dataMap{
											"_gt": int32(63),
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
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithFilterOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with filter on the inner _group selection.",

		Request: `query @explain {
			Author (groupBy: [age]) {
				age
				_group(filter: {age: {_gt: 63}}) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
                     "name": "John Grisham",
                     "age": 65
                 }`,

				`{
                     "name": "Cornelia Funke",
                     "age": 62
                 }`,

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
							"groupByFields": []string{"age"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"docKeys":        nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter": dataMap{
										"age": dataMap{
											"_gt": int32(63),
										},
									},
								},
							},
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
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
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}
