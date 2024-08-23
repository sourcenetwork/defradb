// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mix

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationWithTxnDeletesUserGivenSameTransaction(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create followed by delete in same transaction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(docID: "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotDeletesUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create followed by delete on 2nd transaction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					delete_User(docID: "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesUpdateUserGivenSameTransactions(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Update followed by read in same transaction",
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
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					update_User(input: {age: 28}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotUpdateUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Update followed by read in different transaction",
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
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					update_User(input: {age: 28}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotAllowUpdateInSecondTransactionUser(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Update by two different transactions",
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
					"age": 27
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					update_User(input: {age: 28}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_User(input: {age: 29}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(29),
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction Conflict. Please retry",
			},
			testUtils.Request{
				// Query after transactions have been commited:
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-948fc3eb-9b68-5a8d-9c3c-8f76157002a9",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
