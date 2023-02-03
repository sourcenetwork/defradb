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

func TestExplainDeletionUsingMultiAndSingleIDs_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Explain simple multi-key delete mutation with one key that exists.",

			Request: `mutation @explain {
				delete_author(ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
					_key
				}
			}`,

			Docs: map[int][]string{
				2: {
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
												"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"]) {
					AliasKey: _key
				}
			}`,

			Docs: map[int][]string{
				2: {
					// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
					`{
						"name": "Lone",
						"age":  26,
						"verified": false
					}`,
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
												"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											},
											{
												"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
												"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"]) {
					AliasKey: _key
				}
			}`,

			Docs: map[int][]string{
				2: {
					// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
					`{
						"name": "Lone",
						"age":  26,
						"verified": false
					}`,
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Updates: map[int]map[int][]string{
				0: {
					2: {
						`{
							"age":  28,
							"verified": false
						}`,
					},
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
												"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											},
											{
												"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
												"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(id: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
					_key
				}
			}`,

			Docs: map[int][]string{
				2: {
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
												"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											},
										},
									},
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
		executeTestCase(t, test)
	}
}

func TestExplainDeletionOfDocumentsWithFilter_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Explain deletion using filter - One matching document, that exists.",

			Request: `mutation @explain {
				delete_author(filter: {name: {_eq: "Shahzad"}}) {
					_key
				}
			}`,

			Docs: map[int][]string{
				2: {
					`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Shahzad",
								},
							},
							"ids": []string(nil),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter": dataMap{
											"name": dataMap{
												"_eq": "Shahzad",
											},
										},
										"spans": []dataMap{
											{
												"end":   "/4",
												"start": "/3",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(filter: {
					_and: [
						{age: {_lt: 26}},
						{verified: {_eq: true}},
					]
				}) {
					DeletedKeyByFilter: _key
				}
			}`,

			Docs: map[int][]string{
				2: {
					`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  25,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  6,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  1,
						"verified": true
					}`,
					`{
						"name": "Shahzad Lone",
						"age":  26,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": dataMap{
								"_and": []any{
									dataMap{
										"age": dataMap{
											"_lt": int(26),
										},
									},
									dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
								},
							},
							"ids": []string(nil),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter": dataMap{
											"_and": []any{
												dataMap{
													"age": dataMap{
														"_lt": int(26),
													},
												},
												dataMap{
													"verified": dataMap{
														"_eq": true,
													},
												},
											},
										},
										"spans": []dataMap{
											{
												"end":   "/4",
												"start": "/3",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(filter: {}) {
					DeletedKeyByFilter: _key
				}
			}`,

			Docs: map[int][]string{
				2: {
					`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  25,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  6,
						"verified": true
					}`,
					`{
						"name": "Shahzad",
						"age":  1,
						"verified": true
					}`,
					`{
						"name": "Shahzad Lone",
						"age":  26,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": dataMap{},
							"ids":    []string(nil),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         dataMap{},
										"spans": []dataMap{
											{
												"end":   "/4",
												"start": "/3",
											},
										},
									},
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
		executeTestCase(t, test)
	}
}

func TestExplainDeletionUsingMultiIdsAndSingleIdAndFilter_Failure(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Explain deletion of one document using a list when it doesn't exist.",

			Request: `mutation @explain {
				delete_author(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507e"]) {
					_key
				}
			}`,

			Docs: map[int][]string{
				2: {
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-6a6482a8-24e1-5c73-a237-ca569e41507e",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-6a6482a8-24e1-5c73-a237-ca569e41507f",
												"start": "/3/bae-6a6482a8-24e1-5c73-a237-ca569e41507e",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(ids: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
					_key
				}
			}`,

			Docs: map[int][]string{},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids": []string{
								"bae-028383cc-d6ba-5df7-959f-2bdce3536a05",
								"bae-028383cc-d6ba-5df7-959f-2bdce3536a03",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans": []dataMap{
											{
												"end":   "/3/bae-028383cc-d6ba-5df7-959f-2bdce3536a06",
												"start": "/3/bae-028383cc-d6ba-5df7-959f-2bdce3536a05",
											},
											{
												"end":   "/3/bae-028383cc-d6ba-5df7-959f-2bdce3536a04",
												"start": "/3/bae-028383cc-d6ba-5df7-959f-2bdce3536a03",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(
				    ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "test"],
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
				2: {
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": dataMap{
								"_and": []any{
									dataMap{
										"age": dataMap{
											"_lt": int(26),
										},
									},
									dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
								},
							},
							"ids": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"test",
							},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter": dataMap{
											"_and": []any{
												dataMap{
													"age": dataMap{
														"_lt": int(26),
													},
												},
												dataMap{
													"verified": dataMap{
														"_eq": true,
													},
												},
											},
										},
										"spans": []dataMap{
											{
												"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
												"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											},
											{
												"end":   "/3/tesu",
												"start": "/3/test",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(filter: {name: {_eq: "Shahzad"}}) {
					_key
				}
			}`,

			Docs: map[int][]string{},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Shahzad",
								},
							},
							"ids": []string(nil),
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter": dataMap{
											"name": dataMap{
												"_eq": "Shahzad",
											},
										},
										"spans": []dataMap{
											{
												"end":   "/4",
												"start": "/3",
											},
										},
									},
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

			Request: `mutation @explain {
				delete_author(ids: []) {
					_key
				}
			}`,

			Docs: map[int][]string{
				2: {
					`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"deleteNode": dataMap{
							"filter": nil,
							"ids":    []string{},
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter":         nil,
										"spans":          []dataMap{},
									},
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

			Request: `mutation @explain {
				delete_author(ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"])
			}`,

			Docs: map[int][]string{
				2: {
					// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
					`{
						"name": "Lone",
						"age":  26,
						"verified": false
					}`,
					// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
					`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
				},
			},

			Results: []dataMap{},

			ExpectedError: "Field \"delete_author\" of type \"[author]\" must have a sub selection.",
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
