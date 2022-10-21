// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExplainQueryWithOnlyLimitSpecified(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query Request With Only Limit Specified.",

		Query: `query @explain {
			author(limit: 2) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"limitNode": dataMap{
							"limit":  uint64(2),
							"offset": uint64(0),
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

func TestExplainQueryWithOnlyOffsetSpecified(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query Request With Only Offset Specified.",

		Query: `query @explain {
			author(offset: 2) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"limitNode": dataMap{
							"limit":  nil,
							"offset": uint64(2),
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

func TestExplainQueryWithBothLimitAndOffset(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query Request With Limit and Offset.",

		Query: `query @explain {
			author(limit: 3, offset: 1) {
				name
			}
		}`,

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"limitNode": dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
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

func TestExplainQueryWithOnlyLimitOnChild(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With Only Limit On Child.",

		Query: `query @explain {
			author {
				name
				articles(limit: 1) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"typeIndexJoin": dataMap{
								"joinType": "typeJoinMany",
								"rootName": "author",
								"root": dataMap{
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
								"subTypeName": "articles",
								"subType": dataMap{
									"selectTopNode": dataMap{
										"limitNode": dataMap{
											"limit":  uint64(1),
											"offset": uint64(0),
											"selectNode": dataMap{
												"filter": nil,
												"scanNode": dataMap{
													"collectionID":   "1",
													"collectionName": "article",
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

func TestExplainQueryWithOnlyOffsetOnChild(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With Only Offset On Child.",

		Query: `query @explain {
			author {
				name
				articles(offset: 2) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"typeIndexJoin": dataMap{
								"joinType": "typeJoinMany",
								"rootName": "author",
								"root": dataMap{
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
								"subTypeName": "articles",
								"subType": dataMap{
									"selectTopNode": dataMap{
										"limitNode": dataMap{
											"limit":  nil,
											"offset": uint64(2),
											"selectNode": dataMap{
												"filter": nil,
												"scanNode": dataMap{
													"collectionID":   "1",
													"collectionName": "article",
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

func TestExplainQueryWithBothLimitAndOffsetOnChild(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With Both Limit And Offset On Child.",

		Query: `query @explain {
			author {
				name
				articles(limit: 2, offset: 2) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"typeIndexJoin": dataMap{
								"joinType": "typeJoinMany",
								"rootName": "author",
								"root": dataMap{
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
								"subTypeName": "articles",
								"subType": dataMap{
									"selectTopNode": dataMap{
										"limitNode": dataMap{
											"limit":  uint64(2),
											"offset": uint64(2),
											"selectNode": dataMap{
												"filter": nil,
												"scanNode": dataMap{
													"collectionID":   "1",
													"collectionName": "article",
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

func TestExplainQueryWithLimitOnChildAndBothLimitAndOffsetOnParent(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With Limit On Child And Both Limit And Offset On Parent.",

		Query: `query @explain {
			author(limit: 3, offset: 1) {
				name
				articles(limit: 2) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"limitNode": dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
							"selectNode": dataMap{
								"filter": nil,
								"typeIndexJoin": dataMap{
									"joinType": "typeJoinMany",
									"rootName": "author",
									"root": dataMap{
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
									"subTypeName": "articles",
									"subType": dataMap{
										"selectTopNode": dataMap{
											"limitNode": dataMap{
												"limit":  uint64(2),
												"offset": uint64(0),
												"selectNode": dataMap{
													"filter": nil,
													"scanNode": dataMap{
														"collectionID":   "1",
														"collectionName": "article",
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

func TestExplainQueryWithMultipleConflictingInnerLimits(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With multiple conflicting inner limit nodes.",

		Query: `query @explain {
			author {
				numberOfArts: _count(articles: {})
				articles(limit: 2) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"countNode": dataMap{
							"sources": []dataMap{
								{
									"fieldName": "articles",
									"filter":    nil,
								},
							},
							"selectNode": dataMap{
								"filter": nil,
								"parallelNode": []dataMap{
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"rootName": "author",
											"root": dataMap{
												"scanNode": dataMap{
													"collectionID":   "3",
													"collectionName": "author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/4",
															"start": "/3",
														},
													},
												},
											},
											"subTypeName": "articles",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"limitNode": dataMap{
														"limit":  uint64(2),
														"offset": uint64(0),
														"selectNode": dataMap{
															"filter": nil,
															"scanNode": dataMap{
																"collectionID":   "1",
																"collectionName": "article",
																"filter":         nil,
																"spans": []dataMap{
																	{
																		"end":   "/2",
																		"start": "/1",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"rootName": "author",
											"root": dataMap{
												"scanNode": dataMap{
													"collectionID":   "3",
													"collectionName": "author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/4",
															"start": "/3",
														},
													},
												},
											},
											"subTypeName": "articles",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"filter": nil,
														"scanNode": dataMap{
															"collectionID":   "1",
															"collectionName": "article",
															"filter":         nil,
															"spans": []dataMap{
																{
																	"end":   "/2",
																	"start": "/1",
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainQueryWithMultipleConflictingInnerLimitsAndOuterLimit(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Query With multiple conflicting inner limit nodes and an outer limit.",

		Query: `query @explain {
			author(limit: 3, offset: 1) {
				numberOfArts: _count(articles: {})
				articles(limit: 2) {
					name
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
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

				`{
					"name": "C++ 100",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 101",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 200",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "C++ 202",
					"author_id": "bae-aa839756-588e-5b57-887d-33689a06e375"
				}`,

				`{
					"name": "Rust 100",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 101",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 200",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,

				`{
					"name": "Rust 202",
					"author_id": "bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69"
				}`,
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,

				// _key: bae-aa839756-588e-5b57-887d-33689a06e375
				`{
					"name": "Shahzad Sisley",
					"age": 26,
					"verified": true
				}`,

				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,

				// _key: bae-e7e87bbb-1079-59db-b4b9-0e14b24d5b69
				`{
					"name": "Andrew Lone",
					"age": 28,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"limitNode": dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "articles",
										"filter":    nil,
									},
								},
								"selectNode": dataMap{
									"filter": nil,
									"parallelNode": []dataMap{
										{
											"typeIndexJoin": dataMap{
												"joinType": "typeJoinMany",
												"rootName": "author",
												"root": dataMap{
													"scanNode": dataMap{
														"collectionID":   "3",
														"collectionName": "author",
														"filter":         nil,
														"spans": []dataMap{
															{
																"end":   "/4",
																"start": "/3",
															},
														},
													},
												},
												"subTypeName": "articles",
												"subType": dataMap{
													"selectTopNode": dataMap{
														"limitNode": dataMap{
															"limit":  uint64(2),
															"offset": uint64(0),
															"selectNode": dataMap{
																"filter": nil,
																"scanNode": dataMap{
																	"collectionID":   "1",
																	"collectionName": "article",
																	"filter":         nil,
																	"spans": []dataMap{
																		{
																			"end":   "/2",
																			"start": "/1",
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
										{
											"typeIndexJoin": dataMap{
												"joinType": "typeJoinMany",
												"rootName": "author",
												"root": dataMap{
													"scanNode": dataMap{
														"collectionID":   "3",
														"collectionName": "author",
														"filter":         nil,
														"spans": []dataMap{
															{
																"end":   "/4",
																"start": "/3",
															},
														},
													},
												},
												"subTypeName": "articles",
												"subType": dataMap{
													"selectTopNode": dataMap{
														"selectNode": dataMap{
															"filter": nil,
															"scanNode": dataMap{
																"collectionID":   "1",
																"collectionName": "article",
																"filter":         nil,
																"spans": []dataMap{
																	{
																		"end":   "/2",
																		"start": "/1",
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
