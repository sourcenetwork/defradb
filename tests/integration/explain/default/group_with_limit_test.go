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

func TestDefaultExplainRequestWithLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset on parent groupBy.",

		Request: `query @explain {
			Author(
				groupBy: [name],
				limit: 1,
				offset: 1
			) {
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
						"limitNode": dataMap{
							"limit":  uint64(1),
							"offset": uint64(1),
							"groupNode": dataMap{
								"groupByFields": []string{"name"},
								"childSelects": []dataMap{
									{
										"collectionName": "Author",
										"orderBy":        nil,
										"docKeys":        nil,
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

func TestDefaultExplainRequestWithLimitAndOffsetOnInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset on inner _group selection.",

		Request: `query @explain {
			Author(groupBy: [name]) {
				name
				_group(limit: 2, offset: 1) {
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
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(2),
										"offset": uint64(1),
									},
									"docKeys": nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
								},
							},
							"groupByFields": []string{"name"},
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

func TestDefaultExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset on multiple inner _group selections.",

		Request: `query @explain {
			Author(groupBy: [name]) {
				name
				innerFirstGroup: _group(limit: 1, offset: 2) {
					age
				}
				innerSecondGroup: _group(limit: 2) {
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
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(1),
										"offset": uint64(2),
									},
									"docKeys": nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
								},
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(2),
										"offset": uint64(0),
									},
									"docKeys": nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
								},
							},
							"groupByFields": []string{"name"},
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

func TestDefaultExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with limit and offset on parent groupBy and inner _group selection.",

		Request: `query @explain {
			Author(
				groupBy: [name],
				limit: 1
			) {
				name
				_group(limit: 2) {
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
						"limitNode": dataMap{
							"limit":  uint64(1),
							"offset": uint64(0),
							"groupNode": dataMap{
								"groupByFields": []string{"name"},
								"childSelects": []dataMap{
									{
										"collectionName": "Author",
										"limit": dataMap{
											"limit":  uint64(2),
											"offset": uint64(0),
										},
										"orderBy": nil,
										"docKeys": nil,
										"groupBy": nil,
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
