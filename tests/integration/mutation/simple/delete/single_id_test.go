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
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
	"github.com/stretchr/testify/require"
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
								delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
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
								user(dockey: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
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
						delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
							FancyKey: _key
						}
					}`,

			Results: []map[string]any{
				{
					"FancyKey": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
				},
			},
			ExpectedError: "",
		},
		{
			Description: "Delete an updated document and return an aliased _key name.",
			Request: `mutation {
						delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
							MyTestKey: _key
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
					"MyTestKey": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
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
					delete_user(id: "bae-028383cc-d6ba-5df7-959f-2bdce3536a05") {
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
					delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca811") {
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
						delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810")
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
			ExpectedError: "Field \"delete_user\" of type \"[user]\" must have a sub selection.",
		},

		{
			Description: "Deletion of a document without _key sub-selection.",
			Request: `mutation {
						delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
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
			ExpectedError: "Syntax Error GraphQL request (2:67) Unexpected empty IN {}\n\n1: mutation {\n2: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009delete_user(id: \"bae-8ca944fd-260e-5a44-b88f-326d9faca810\") {\n                                                                     ^\n3: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009}\n",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeletionOfADocumentUsingSingleKeyWithShowDeletedDocumentQuery_Success(t *testing.T) {
	jsonString := `{
		"name": "John",
		"age": 43
	}`
	doc, err := client.NewDocFromJSON([]byte(jsonString))
	require.NoError(t, err)

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          jsonString,
			},
			testUtils.Request{
				Request: fmt.Sprintf(`mutation {
						delete_User(id: "%s") {
							_status
							_key
						}
					}`, doc.Key()),
				Results: []map[string]any{
					{
						// Note: This should show a `Deleted` status but the order of the planNodes
						// makes it so the status is requested prior to deleting. If the planNode ordering
						// can be altered, this can change in the future.
						"_status": "Active",
						"_key":    doc.Key().String(),
					},
				},
			},
			testUtils.Request{
				Request: `query {
						User(showDeleted: true) {
							_status
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"_status": "Deleted",
						"name":    "John",
						"age":     uint64(43),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"User"}, test)
}
