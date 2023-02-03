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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExplainGroupByWithGroupLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "Explain query with limit and offset on parent groupBy.",

		Request: `query @explain {
			author(
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

		Results: []dataMap{
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
										"collectionName": "author",
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
										"collectionName": "author",
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

	executeTestCase(t, test)
}

func TestExplainGroupByWithGroupLimitAndOffsetOnChild(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "Explain query with limit and offset on child groupBy.",

		Request: `query @explain {
			author(groupBy: [name]) {
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "author",
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
									"collectionName": "author",
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

	executeTestCase(t, test)
}

func TestExplainGroupByWithGroupLimitOnChildMultipleRendered(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with limit on child groupBy (multiple rendered).",

		Request: `query @explain {
			author(groupBy: [name]) {
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"groupNode": dataMap{
							"childSelects": []dataMap{
								{
									"collectionName": "author",
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
									"collectionName": "author",
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
									"collectionName": "author",
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

	executeTestCase(t, test)
}

func TestExplainGroupByWithGroupLimitOnParentAndChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with limit on parent and child groupBy.",

		Request: `query @explain {
			author(
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

		Results: []dataMap{
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
										"collectionName": "author",
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
										"collectionName": "author",
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

	executeTestCase(t, test)
}
