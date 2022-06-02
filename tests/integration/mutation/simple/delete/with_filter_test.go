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

func TestDeletionOfDocumentsWithFilter_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Delete using filter - One matching document, that exists.",

			Request: `mutation {
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

			Results: []map[string]interface{}{
				{
					"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Delete using filter - Multiple matching documents that exist.",
			Request: `mutation {
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

			Results: []map[string]interface{}{
				{
					"_key": "bae-4b5b1765-560c-5843-9abc-24d21d8aa598",
				},
				{
					"_key": "bae-5a8530c0-c521-5e83-8243-4ce267bc76fa",
				},
				{
					"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
				{
					"_key": "bae-ca88bc10-1415-59b1-a72c-d19ed44d4e15",
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Delete using filter - Multiple matching documents that exist with alias.",

			Request: `mutation {
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

			Results: []map[string]interface{}{
				{
					"DeletedKeyByFilter": "bae-4b5b1765-560c-5843-9abc-24d21d8aa598",
				},
				{
					"DeletedKeyByFilter": "bae-5a8530c0-c521-5e83-8243-4ce267bc76fa",
				},
				{
					"DeletedKeyByFilter": "bae-ca88bc10-1415-59b1-a72c-d19ed44d4e15",
				},
			},

			ExpectedError: "",
		},

		{
			Description: "Delete using filter - Match everything in this collection.",

			Request: `mutation {
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

			Results: []map[string]interface{}{
				{
					"DeletedKeyByFilter": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
				{
					"DeletedKeyByFilter": "bae-4b5b1765-560c-5843-9abc-24d21d8aa598",
				},
				{
					"DeletedKeyByFilter": "bae-5a8530c0-c521-5e83-8243-4ce267bc76fa",
				},
				{
					"DeletedKeyByFilter": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
				{
					"DeletedKeyByFilter": "bae-ca88bc10-1415-59b1-a72c-d19ed44d4e15",
				},
			},

			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeletionOfDocumentsWithFilter_Failure(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "No delete with filter: because no document matches filter.",

			Request: `mutation {
						delete_user(filter: {name: {_eq: "Lone"}}) {
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

			Results: []map[string]interface{}{},

			ExpectedError: "",
		},

		{
			Description: "No delete with filter: because the collection is empty.",

			Request: `mutation {
						delete_user(filter: {name: {_eq: "Shahzad"}}) {
							_key
						}
					}`,

			Docs: map[int][]string{},

			Results: []map[string]interface{}{},

			ExpectedError: "",
		},

		{
			Description: "No delete with filter: because has no sub-selection.",

			Request: `mutation {
						delete_user(filter: {name: {_eq: "Shahzad"}})
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

			Results: []map[string]interface{}{},

			ExpectedError: "[Field \"delete_user\" of type \"[user]\" must have a sub selection.]",
		},

		{
			Description: "No delete with filter: because has no _key in sub-selection.",

			Request: `mutation {
						delete_user(filter: {name: {_eq: "Shahzad"}}) {
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

			Results: []map[string]interface{}{},

			ExpectedError: "Syntax Error GraphQL request (2:53) Unexpected empty IN {}\n\n1: mutation {\n2: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009delete_user(filter: {name: {_eq: \"Shahzad\"}}) {\n                                                       ^\n3: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009}\n",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}
