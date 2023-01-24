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

func TestExplainGroupByWithFilterOnParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a grouping with filter on parent.",

		Request: `query @explain {
			author (
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"groupByFields": []string{"age"},
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
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter": dataMap{
										"age": dataMap{
											"_gt": int(63),
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

	executeTestCase(t, test)
}

func TestExplainGroupByWithFilterOnInnerGroupSelection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a grouping with filter on the inner group selection.",

		Request: `query @explain {
			author (groupBy: [age]) {
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"groupByFields": []string{"age"},
							"childSelects": []dataMap{
								{
									"collectionName": "author",
									"docKeys":        nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
									"filter": dataMap{
										"age": dataMap{
											"_gt": int(63),
										},
									},
								},
							},
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"filter":         nil,
									"collectionID":   "3",
									"collectionName": "author",
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

	executeTestCase(t, test)
}
