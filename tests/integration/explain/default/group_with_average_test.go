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

func TestExplainGroupByWithAverageOnAnInnerField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a groupBy with average on an field.",

		Request: `query @explain {
			author (groupBy: [name]) {
				name
				_avg(_group: {field: age})
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
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "_group",
										"filter": dataMap{
											"age": dataMap{
												"_ne": nil,
											},
										},
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "age",
											"fieldName":      "_group",
											"filter": dataMap{
												"age": dataMap{
													"_ne": nil,
												},
											},
										},
									},
									"groupNode": dataMap{
										"childSelects": []dataMap{
											{
												"collectionName": "author",
												"docKeys":        nil,
												"groupBy":        nil,
												"limit":          nil,
												"orderBy":        nil,
												"filter": dataMap{
													"age": dataMap{
														"_ne": nil,
													},
												},
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainGroupByWithAnAverageInsideTheInnerGroupOnAField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a groupBy with average of inside the inner group (on a field).",

		Request: `query @explain {
			author (groupBy: [name]) {
				name
				_avg(_group: {field: _avg})
				_group(groupBy: [verified]) {
					verified
					_avg(_group: {field: age})
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
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "_group",
										"filter":    nil,
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "_avg",
											"fieldName":      "_group",
											"filter":         nil,
										},
									},
									"groupNode": dataMap{
										"childSelects": []dataMap{
											{
												"collectionName": "author",
												"groupBy":        []string{"verified", "name"},
												"docKeys":        nil,
												"filter":         nil,
												"limit":          nil,
												"orderBy":        nil,
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainGroupByWithAnAverageInsideTheInnerGroupAndNestedGroupBy(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a groupBy with average of inside the inner group with nested groupBy.",

		Request: `query @explain {
			author (groupBy: [name]) {
				name
				_avg(_group: {field: _avg})
				_group(groupBy: [verified]) {
					verified
						_avg(_group: {field: age})
						_group (groupBy: [age]){
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "_group",
										"filter":    nil,
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "_avg",
											"fieldName":      "_group",
											"filter":         nil,
										},
									},
									"groupNode": dataMap{
										"childSelects": []dataMap{
											{
												"collectionName": "author",
												"groupBy":        []string{"verified", "name"},
												"docKeys":        nil,
												"filter":         nil,
												"limit":          nil,
												"orderBy":        nil,
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainGroupByWihAnAverageInsideTheInnerGroupAndNestedGroupByWithAnAverage(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a groupBy with average of inside the inner group with nested groupBy with and average.",

		Request: `query @explain {
			author (groupBy: [name]) {
				name
				_avg(_group: {field: _avg})
				_group(groupBy: [verified]) {
					verified
						_avg(_group: {field: age})
						_group (groupBy: [age]){
							age
							_avg(_group: {field: age})
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

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "_group",
										"filter":    nil,
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "_avg",
											"fieldName":      "_group",
											"filter":         nil,
										},
									},
									"groupNode": dataMap{
										"childSelects": []dataMap{
											{
												"collectionName": "author",
												"groupBy":        []string{"verified", "name"},
												"docKeys":        nil,
												"filter":         nil,
												"limit":          nil,
												"orderBy":        nil,
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
				},
			},
		},
	}

	executeTestCase(t, test)
}
