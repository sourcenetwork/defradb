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

func TestDeletionOfADocumentUsingSingleKey_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Simple delete mutation where one element exists.",
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`,
				},
			},
			TransactionalRequests: []testUtils.TransactionRequest{
				{
					TransactionId: 0,
					Request: `mutation {
								delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
									_key
								}
							}`,
					Results: []map[string]any{
						{
							"_key": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
						},
					},
				},
				{
					TransactionId: 0,
					Request: `query {
								User(dockey: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
									_key
								}
							}`,

					// explicitly empty
					Results: []map[string]any{},
				},
			},
		},

		{
			Description: "Simple delete mutation with an aliased _key name.",
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`,
				},
			},
			Request: `mutation {
						delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
							fancyKey: _key
						}
					}`,

			Results: []map[string]any{
				{
					"fancyKey": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
				},
			},
			ExpectedError: "",
		},
		{
			Description: "Delete an updated document and return an aliased _key name.",
			Request: `mutation {
						delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
							myTestKey: _key
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`,
				},
			},
			Updates: map[int]map[int][]string{
				0: {
					0: {
						`{
							"age":  27,
							"points": 48.2,
							"verified": false
						}`,
					},
				},
			},
			Results: []map[string]any{
				{
					"myTestKey": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
				},
			},
			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeleteWithUnknownIdEmptyCollection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion using id that doesn't exist, where the collection is empty.",
		Request: `mutation {
					delete_User(id: "bae-028383cc-d6ba-5df7-959f-2bdce3536a05") {
						_key
					}
				}`,
		Docs:    map[int][]string{},
		Results: []map[string]any{},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeleteWithUnknownId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion using id that doesn't exist, where the collection is non-empty.",
		Request: `mutation {
					delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca811") {
						_key
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"age":  26,
					"points": 48.5,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeletionOfADocumentUsingSingleKey_Failure(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Deletion of a document without sub selection, should give error.",
			Request: `mutation {
						delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810")
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`,
				},
			},
			Results:       []map[string]any{},
			ExpectedError: "Field \"delete_User\" of type \"[User]\" must have a sub selection.",
		},

		{
			Description: "Deletion of a document without _key sub-selection.",
			Request: `mutation {
						delete_User(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`,
				},
			},
			Results:       []map[string]any{},
			ExpectedError: "Syntax Error GraphQL request (2:67) Unexpected empty IN {}\n\n1: mutation {\n2: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009delete_User(id: \"bae-8ca944fd-260e-5a44-b88f-326d9faca810\") {\n                                                                     ^\n3: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009}\n",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}
