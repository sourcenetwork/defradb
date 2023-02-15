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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

func TestMutationWithTxnDeletesUserGivenSameTransaction(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Create followed by delete in same transaction",
		TransactionalRequests: []testUtils.TransactionRequest{
			{
				TransactionId: 0,
				Request: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 0,
				Request: `mutation {
					delete_user(id: "bae-88b63198-7d38-5714-a9ff-21ba46374fd1") {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotDeletesUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Create followed by delete on 2nd transaction",
		TransactionalRequests: []testUtils.TransactionRequest{
			{
				TransactionId: 0,
				Request: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 1,
				Request: `mutation {
					delete_user(id: "bae-88b63198-7d38-5714-a9ff-21ba46374fd1") {
						_key
					}
				}`,
				Results: []map[string]any{},
			},
			{
				TransactionId: 0,
				Request: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(27),
					},
				},
			},
			{
				TransactionId: 1,
				Request: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesUpdateUserGivenSameTransactions(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Update followed by read in same transaction",
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27
				}`,
			},
		},
		TransactionalRequests: []testUtils.TransactionRequest{
			{
				TransactionId: 0,
				Request: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 0,
				Request: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotUpdateUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Update followed by read in different transaction",
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27
				}`,
			},
		},
		TransactionalRequests: []testUtils.TransactionRequest{
			{
				TransactionId: 0,
				Request: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
			{
				TransactionId: 1,
				Request: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(27),
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotAllowUpdateInSecondTransactionUser(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Update by two different transactions",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.TransactionRequest2{
				TransactionId: 0,
				Request: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
			testUtils.TransactionRequest2{
				TransactionId: 1,
				Request: `mutation {
					update_user(data: "{\"age\": 29}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(29),
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionId: 0,
			},
			testUtils.TransactionCommit{
				TransactionId: 1,
				ExpectedError: "Transaction Conflict. Please retry",
			},
			testUtils.Request{
				// Query after transactions have been commited:
				Request: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
		},
	}

	simpleTests.Execute(t, test)
}
