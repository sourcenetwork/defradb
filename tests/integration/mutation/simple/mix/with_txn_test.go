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
	test := testUtils.QueryTestCase{
		Description: "Create followed by delete in same transaction",
		TransactionalQueries: []testUtils.TransactionQuery{
			{
				TransactionId: 0,
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 0,
				Query: `mutation {
					delete_user(id: "bae-88b63198-7d38-5714-a9ff-21ba46374fd1") {
						_key
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
		},
		// Map store does not support transactions
		DisableMapStore: true,
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotDeletesUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Create followed by delete on 2nd transaction",
		TransactionalQueries: []testUtils.TransactionQuery{
			{
				TransactionId: 0,
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 1,
				Query: `mutation {
					delete_user(id: "bae-88b63198-7d38-5714-a9ff-21ba46374fd1") {
						_key
					}
				}`,
				Results: []map[string]interface{}{},
			},
			{
				TransactionId: 0,
				Query: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(27),
					},
				},
			},
			{
				TransactionId: 1,
				Query: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{},
			},
		},
		// Map store does not support transactions
		DisableMapStore: true,
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesUpdateUserGivenSameTransactions(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Update followed by read in same transaction",
		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27
			}`)},
		},
		TransactionalQueries: []testUtils.TransactionQuery{
			{
				TransactionId: 0,
				Query: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
					},
				},
			},
			{
				TransactionId: 0,
				Query: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
		},
		// Map store does not support transactions
		DisableMapStore: true,
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotUpdateUserGivenDifferentTransactions(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Update followed by read in different transaction",
		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27
			}`)},
		},
		TransactionalQueries: []testUtils.TransactionQuery{
			{
				TransactionId: 0,
				Query: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
			{
				TransactionId: 1,
				Query: `query {
					user {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(27),
					},
				},
			},
		},
		// Map store does not support transactions
		DisableMapStore: true,
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationWithTxnDoesNotAllowUpdateInSecondTransactionUser(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Update by two different transactions",
		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27
			}`)},
		},
		TransactionalQueries: []testUtils.TransactionQuery{
			{
				TransactionId: 0,
				Query: `mutation {
					update_user(data: "{\"age\": 28}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(28),
					},
				},
			},
			{
				TransactionId: 1,
				Query: `mutation {
					update_user(data: "{\"age\": 29}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]interface{}{
					{
						"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name": "John",
						"age":  uint64(29),
					},
				},
				ExpectedError: "Transaction Conflict. Please retry",
			},
		},
		// Query after transactions have been commited:
		Query: `query {
			user {
				_key
				name
				age
			}
		}`,
		Results: []map[string]interface{}{
			{
				"_key": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
				"name": "John",
				"age":  uint64(28),
			},
		},
		// Map store does not support transactions
		DisableMapStore: true,
	}

	simpleTests.ExecuteTestCase(t, test)
}
