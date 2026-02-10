// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package conflict

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestTxnConflict_UpdateVsDelete tests that when two concurrent transactions
// target the same document — one updating and one deleting — the second
// transaction to commit fails with a transaction conflict.
func TestTxnConflict_UpdateVsDelete(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.CreateDoc{
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction conflict",
			},
			// Verify the update from transaction 0 took effect.
			&action.Request{
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

// TestTxnConflict_DeleteVsUpdate tests that when two concurrent transactions
// target the same document — one deleting and one updating — the second
// transaction to commit fails with a transaction conflict.
func TestTxnConflict_DeleteVsUpdate(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27
				}`,
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
			&action.Request{
				TransactionID: immutable.Some(1),
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction conflict",
			},
			// Verify the delete from transaction 0 took effect.
			&action.Request{
				Request: `query {
					User {
						_docID
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

// TestTxnConflict_DeleteVsDelete tests that when two concurrent transactions
// both attempt to delete the same document, the second transaction to commit
// fails with a transaction conflict.
func TestTxnConflict_DeleteVsDelete(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27
				}`,
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
			&action.Request{
				TransactionID: immutable.Some(1),
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction conflict",
			},
			// Verify the document was deleted by transaction 0.
			&action.Request{
				Request: `query {
					User {
						_docID
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

// TestTxnConflict_CreateVsCreate tests that when two concurrent transactions
// both create a document with identical content (producing the same content-addressed
// document ID), the second transaction to commit fails with a transaction conflict.
func TestTxnConflict_CreateVsCreate(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			// Both transactions create a document with identical content. Because document
			// IDs are content-addressed, both produce the same docID, causing a conflict.
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
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
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction conflict",
			},
			// Verify only one document exists (created by transaction 0).
			&action.Request{
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
