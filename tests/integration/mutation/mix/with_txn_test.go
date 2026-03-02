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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationWithTxnDeletesUserGivenSameTransaction(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					add_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"add_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
						},
					},
				},
			},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(docID: "bae-bb8ed746-4570-5651-ac69-39a21f733211") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
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
		// LevelDB does not support concurrent transactions
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					add_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"add_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
						},
					},
				},
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					delete_User(docID: "bae-bb8ed746-4570-5651-ac69-39a21f733211") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{},
				},
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					update_User(input: {age: 28}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
						},
					},
				},
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
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
		// LevelDB does not support concurrent transactions
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
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
		// LevelDB does not support concurrent transactions
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
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
				ExpectedError: "transaction conflict. Please retry",
			},
			&action.Request{
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
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
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
