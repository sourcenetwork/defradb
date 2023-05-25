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

func TestDefaultExplainRequestWithDescendingOrderOnParentGroupBy(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with order (descending) on parent groupBy.",

		Request: `query @explain {
			Author(groupBy: [name], order: {name: DESC}) {
				name
				_group {
					age
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 65
				}`,
				`{
					"name": "John Grisham",
					"verified": false,
					"age": 2
				}`,
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 50
				}`,
				`{
					"name": "Cornelia Funke",
					"verified": true,
					"age": 62
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
			},
		},

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields":    []string{"name"},
								},
							},
							"groupNode": dataMap{
								"groupByFields": []string{"name"},
								"childSelects": []dataMap{
									{
										"collectionName": "Author",
										"docKeys":        nil,
										"orderBy":        nil,
										"groupBy":        nil,
										"limit":          nil,
										"filter":         nil,
									},
								},
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
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithAscendingOrderOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with order (ascending) on inner _group selection.",

		Request: `query @explain {
			Author(groupBy: [name]) {
				name
				_group (order: {age: ASC}){
					age
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 65
				}`,
				`{
					"name": "John Grisham",
					"verified": false,
					"age": 2
				}`,
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 50
				}`,
				`{
					"name": "Cornelia Funke",
					"verified": true,
					"age": 62
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
			},
		},

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"orderBy": []dataMap{
										{
											"direction": "ASC",
											"fields":    []string{"age"},
										},
									},
									"docKeys": nil,
									"groupBy": nil,
									"limit":   nil,
									"filter":  nil,
								},
							},
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

func TestDefaultExplainRequestWithOrderOnParentGroupByAndOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with order on parent groupBy and inner _group selection.",

		Request: `query @explain {
			Author(groupBy: [name], order: {name: DESC}) {
				name
				_group (order: {age: ASC}){
					age
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 65
				}`,
				`{
					"name": "John Grisham",
					"verified": false,
					"age": 2
				}`,
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 50
				}`,
				`{
					"name": "Cornelia Funke",
					"verified": true,
					"age": 62
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
			},
		},

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields":    []string{"name"},
								},
							},
							"groupNode": dataMap{
								"groupByFields": []string{"name"},
								"childSelects": []dataMap{
									{
										"collectionName": "Author",
										"orderBy": []dataMap{
											{
												"direction": "ASC",
												"fields":    []string{"age"},
											},
										},
										"docKeys": nil,
										"groupBy": nil,
										"limit":   nil,
										"filter":  nil,
									},
								},
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
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with order on nested parent groupBy and on nested parent's inner _group.",

		Request: `query @explain {
			Author(groupBy: [name]) {
				name
				_group (
					groupBy: [verified],
					order: {verified: ASC}
				){
					verified
					_group (order: {age: DESC}) {
						age
					}
				}
			}
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 65
				}`,
				`{
					"name": "John Grisham",
					"verified": false,
					"age": 2
				}`,
				`{
					"name": "John Grisham",
					"verified": true,
					"age": 50
				}`,
				`{
					"name": "Cornelia Funke",
					"verified": true,
					"age": 62
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
				`{
					"name": "Twin",
					"verified": true,
					"age": 63
				}`,
			},
		},

		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"orderBy": []dataMap{
										{
											"direction": "ASC",
											"fields":    []string{"verified"},
										},
									},
									"groupBy": []string{"verified", "name"},
									"docKeys": nil,
									"limit":   nil,
									"filter":  nil,
								},
							},
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
