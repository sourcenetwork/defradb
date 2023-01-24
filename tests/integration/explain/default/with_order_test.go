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

func TestExplainAscendingOrderQueryOnParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain An Ascending Order Query On Parent Field.",

		Query: `query @explain {
			author(order: {age: ASC}) {
				name
				age
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
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
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
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields": []string{
										"age",
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

func TestExplainQueryWithMultiOrderFieldsOnParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain Query With Multiple Order Fields on the Parent.",

		Query: `query @explain {
			author(order: {name: ASC, age: DESC}) {
				name
				age
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
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
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
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields": []string{
										"name",
									},
								},
								{
									"direction": "DESC",
									"fields": []string{
										"age",
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

func TestExplainQueryWithOrderFieldOnChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain Query With Order Field On A Child.",

		Query: `query @explain {
			author {
				name
				articles(order: {name: DESC}) {
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
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
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
										"orderNode": dataMap{
											"orderings": []dataMap{
												{
													"direction": "DESC",
													"fields": []string{
														"name",
													},
												},
											},
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

func TestExplainQueryWithOrderOnBothTheParentAndChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain A Query With Order On Parent and An Order on Child.",

		Query: `query @explain {
			author(order: {name: ASC}) {
				name
				articles(order: {name: DESC}) {
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
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields": []string{
										"name",
									},
								},
							},
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
											"orderNode": dataMap{
												"orderings": []dataMap{
													{
														"direction": "DESC",
														"fields": []string{
															"name",
														},
													},
												},
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

func TestExplainQueryWhereParentIsOrderedByChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain Query Where The Parent Is Ordered By It's Child.",

		Query: `query @explain {
			author(
				order: {
					articles: {name: ASC}
				}
			) {
				articles {
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
			},

			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"orderNode": dataMap{
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields": []string{
										"articles",
										"name",
									},
								},
							},
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
