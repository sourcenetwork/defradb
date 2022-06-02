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

func TestDeletionOfMultipleDocumentUsingMultipleKeys_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Simple multi-key delete mutation with one key that exists.",
			Request: `mutation {
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
			Results: []map[string]interface{}{
				{
					"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Delete multiple documents that exist, when given multiple keys.",
			Request: `mutation {
						delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
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
					(`{
						"name": "John",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`),
				},
			},
			Results: []map[string]interface{}{
				{
					"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
				{
					"_key": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Delete multiple documents that exist, when given multiple keys with alias.",
			Request: `mutation {
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
			Results: []map[string]interface{}{
				{
					"AliasKey": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
				{
					"AliasKey": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Delete multiple documents that exist, where an update happens too.",
			Request: `mutation {
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
			Results: []map[string]interface{}{
				{
					"AliasKey": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
				{
					"AliasKey": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
			},
			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeletionOfMultipleDocumentUsingMultipleKeys_Failure(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Deletion of one document using a list when it doesn't exist, in a non-empty collection.",
			Request: `mutation {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "No document for the given key exists",
		},

		{
			Description: "Simple multi-key delete mutation while no documents exist.",
			Request: `mutation {
						delete_user(ids: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
							_key
						}
					}`,
			Docs:          map[int][]string{},
			Results:       []map[string]interface{}{},
			ExpectedError: "No document for the given key exists",
		},

		{
			Description: "Simple multi-key delete mutation while one document doesn't exist.",
			Request: `mutation {
						delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "No document for the given key exists",
		},

		{
			Description: "Simple multi-key delete used with filter.",
			Request: `mutation {
						delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"], filter: {}) {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "Error: can't use filter and id / ids together.",
		},

		{
			Description: "Simple multi-key delete mutation but no ids given.",
			Request: `mutation {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "Error: no id(s) provided while delete mutation.",
		},

		{
			Description: "Simple multi-key delete mutation but no ids given.",
			Request: `mutation {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "Error: no id(s) provided while delete mutation.",
		},

		{
			Description: "Delete multiple documents that exist without sub selection, should give error.",
			Request: `mutation {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "[Field \"delete_user\" of type \"[user]\" must have a sub selection.]",
		},

		{
			Description: "Delete multiple documents that exist without _key sub-selection.",
			Request: `mutation {
						delete_user(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
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
			Results:       []map[string]interface{}{},
			ExpectedError: "Syntax Error GraphQL request (2:114) Unexpected empty IN {}\n\n1: mutation {\n2: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009delete_user(ids: [\"bae-6a6482a8-24e1-5c73-a237-ca569e41507d\", \"bae-3a1a496e-24eb-5ae3-9c17-524c146a393e\"]) {\n                                                                                                                    ^\n3: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009}\n",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}
