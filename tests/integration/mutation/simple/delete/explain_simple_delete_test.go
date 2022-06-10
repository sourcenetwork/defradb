// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

type dataMap = map[string]interface{}

func TestExplainDeletionUsingMultiAndSingleIDs_Success(t *testing.T) {
	tests := []testUtils.QueryTestCase{

		{
			Description: "Explain simple multi-key delete mutation with one key that exists.",

			Query: `mutation @explain {
								delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"]) {
									_key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain deletion of multiple documents that exist, when given multiple keys with alias.",

			Query: `mutation @explain {
								delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
									AliasKey: _key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "John",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
											"bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain the deletion of multiple documents that exist, where an update has happened too.",

			Query: `mutation @explain {
								delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
									AliasKey: _key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "John",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Updates: map[int][]string{
				0: {
					(`{
										"age":  27,
										"points": 48.2,
										"verified": false
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
											"bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain simple delete mutation with single id, where the doc exists.",

			Query: `mutation @explain {
								delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
									_key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.5,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-8ca944fd-260e-5a44-b88f-326d9faca810",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestExplainDeletionOfDocumentsWithFilter_Success(t *testing.T) {
	tests := []testUtils.QueryTestCase{

		{
			Description: "Explain deletion using filter - One matching document, that exists.",

			Query: `mutation @explain {
						delete_user(filter: {name: {_eq: "Shahzad"}}) {
							_key
						}
					}`,

			Docs: map[int][]string{
				0: {
					(`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": dataMap{
											"name": dataMap{
												"$eq": "Shahzad",
											},
										},
										"ids": []string(nil),
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain deletion using filter - Multiple matching documents that exist with alias.",

			Query: `mutation @explain {
								delete_user(filter: {
									_and: [
										{age: {_lt: 26}},
										{verified: {_eq: true}},
									]
								}) {
									DeletedKeyByFilter: _key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  25,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  6,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  1,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "John",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": dataMap{
											"$and": []interface{}{
												dataMap{
													"age": dataMap{
														"$lt": int64(26),
													},
												},
												dataMap{
													"verified": dataMap{
														"$eq": true,
													},
												},
											},
										},
										"ids": []string(nil),
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain deletion using filter - Match everything in this collection.",

			Query: `mutation @explain {
								delete_user(filter: {}) {
									DeletedKeyByFilter: _key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  25,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  6,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "Shahzad",
								"age":  1,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "John",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": dataMap{},
										"ids":    []string(nil),
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestExplainDeletionUsingMultiIdsAndSingleIdAndFilter_Failure(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Explain deletion of one document using a list when it doesn't exist.",

			Query: `mutation @explain {
						delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507e"]) {
							_key
						}
					}`,

			Docs: map[int][]string{
				0: {
					(`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-6a6482a8-24e1-5c73-a237-ca569e41507e",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain a simple multi-key delete mutation while no documents exist.",

			Query: `mutation @explain {
								delete_user(ids: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
									_key
								}
							}`,
			Docs: map[int][]string{},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids": []string{
											"bae-028383cc-d6ba-5df7-959f-2bdce3536a05",
											"bae-028383cc-d6ba-5df7-959f-2bdce3536a03",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain a simple multi-key delete used with filter.",

			Query: `mutation @explain {
								delete_user(
								    ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "test"],
								    filter: {
									    _and: [
									    	{age: {_lt: 26}},
									    	{verified: {_eq: true}},
									    ]
									}
								) {
									_key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": dataMap{
											"$and": []interface{}{
												dataMap{
													"age": dataMap{
														"$lt": int64(26),
													},
												},
												dataMap{
													"verified": dataMap{
														"$eq": true,
													},
												},
											},
										},
										"ids": []string{
											"bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
											"test",
										},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain no delete with filter: because the collection is empty.",

			Query: `mutation @explain {
						delete_user(filter: {name: {_eq: "Shahzad"}}) {
							_key
						}
					}`,

			Docs: map[int][]string{},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": dataMap{
											"name": dataMap{
												"$eq": "Shahzad",
											},
										},
										"ids": []string(nil),
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain a simple multi-key delete mutation but no ids given.",

			Query: `mutation @explain {
								delete_user(ids: []) {
									_key
								}
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"renderNode": dataMap{
								"selectNode": dataMap{
									"deleteNode": dataMap{
										"filter": nil,
										"ids":    []string{},
									},
									"filter": nil,
								},
							},
						},
					},
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Explain deletion of multiple documents that exist without sub selection, should give error.",

			Query: `mutation @explain {
								delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"])
							}`,

			Docs: map[int][]string{
				0: {
					(`{
								"name": "Shahzad",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
					(`{
								"name": "John",
								"age":  26,
								"points": 48.48,
								"verified": true
							}`),
				},
			},

			Results: []dataMap{},

			ExpectedError: "[Field \"delete_user\" of type \"[user]\" must have a sub selection.]",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}
