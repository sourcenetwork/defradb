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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

func TestDeletionOfMultipleDocumentUsingMultipleKeysWhereOneExists(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple multi-key delete mutation with one key that exists.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
						points: Float
						verified: Boolean
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"]) {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User(dockeys: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"]) {
						_key
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeletionOfMultipleDocumentUsingMultipleKeys_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Delete multiple documents that exist, when given multiple keys.",
			Request: `mutation {
						delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
							_key
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
					`{
						"name": "John",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
				{
					"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Delete multiple documents that exist, when given multiple keys with alias.",
			Request: `mutation {
						delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
							AliasKey: _key
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
					`{
						"name": "John",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
				},
			},
			Results: []map[string]any{
				{
					"AliasKey": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
				{
					"AliasKey": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Delete multiple documents that exist, where an update happens too.",
			Request: `mutation {
						delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
							AliasKey: _key
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
					`{
						"name": "John",
						"age":  26,
						"points": 48.48,
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
					"AliasKey": "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e",
				},
				{
					"AliasKey": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
				},
			},
			ExpectedError: "",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeleteWithEmptyIdsSet(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion of using ids, empty ids set.",
		Request: `mutation {
					delete_User(ids: []) {
						_key
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeleteWithSingleUnknownIds(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion of using ids, single unknown item.",
		Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507e"]) {
						_key
					}
				}`,
		Results: []map[string]any{},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeleteWithMultipleUnknownIds(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion of using ids, multiple unknown items.",
		Request: `mutation {
					delete_User(ids: ["bae-028383cc-d6ba-5df7-959f-2bdce3536a05", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
						_key
					}
				}`,
		Results: []map[string]any{},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeleteWithUnknownAndKnownIds(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion of using ids, known and unknown items.",
		Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-028383cc-d6ba-5df7-959f-2bdce3536a03"]) {
						_key
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeleteWithKnownIdsAndEmptyFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Deletion of using ids and filter, known id and empty filter.",
		Request: `mutation {
					delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d"], filter: {}) {
						_key
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key": "bae-6a6482a8-24e1-5c73-a237-ca569e41507d",
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestDeletionOfMultipleDocumentUsingMultipleKeys_Failure(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Delete multiple documents that exist without sub selection, should give error.",
			Request: `mutation {
						delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"])
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
					`{
						"name": "John",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
				},
			},
			Results:       []map[string]any{},
			ExpectedError: "Field \"delete_User\" of type \"[User]\" must have a sub selection.",
		},

		{
			Description: "Delete multiple documents that exist without _key sub-selection.",
			Request: `mutation {
						delete_User(ids: ["bae-6a6482a8-24e1-5c73-a237-ca569e41507d", "bae-3a1a496e-24eb-5ae3-9c17-524c146a393e"]) {
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
					`{
						"name": "John",
						"age":  26,
						"points": 48.48,
						"verified": true
					}`,
				},
			},
			Results:       []map[string]any{},
			ExpectedError: "Syntax Error GraphQL request (2:114) Unexpected empty IN {}\n\n1: mutation {\n2: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009delete_User(ids: [\"bae-6a6482a8-24e1-5c73-a237-ca569e41507d\", \"bae-3a1a496e-24eb-5ae3-9c17-524c146a393e\"]) {\n                                                                                                                    ^\n3: \\u0009\\u0009\\u0009\\u0009\\u0009\\u0009}\n",
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}

func TestDeletionOfMultipleDocumentsUsingSingleKeyWithShowDeletedDocumentQuery_Success(t *testing.T) {
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
				Doc: `{
					"name": "John",
					"age": 43
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Andy",
					"age": 74
				}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_User(ids: ["bae-05de0e64-f300-55b3-8973-5fa79045a083", "bae-07e5c44c-ee88-5c92-85ad-fb3148c48bef"]){
							_key
						}
					}`,
				Results: []map[string]any{
					{
						"_key": "bae-05de0e64-f300-55b3-8973-5fa79045a083",
					},
					{
						"_key": "bae-07e5c44c-ee88-5c92-85ad-fb3148c48bef",
					},
				},
			},
			testUtils.Request{
				Request: `query {
						User(showDeleted: true) {
							_deleted
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"_deleted": true,
						"name":     "Andy",
						"age":      uint64(74),
					},
					{
						"_deleted": true,
						"name":     "John",
						"age":      uint64(43),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeletionOfMultipleDocumentsUsingEmptySet(t *testing.T) {
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
				Doc: `{
					"name": "John",
					"age": 43
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Andy",
					"age": 74
				}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_User(ids: []){
							_key
						}
					}`,
				Results: []map[string]any{},
			},
			testUtils.Request{
				// Make sure no documents have been deleted
				Request: `query {
						User {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Andy",
						"age":  uint64(74),
					},
					{
						"name": "John",
						"age":  uint64(43),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
