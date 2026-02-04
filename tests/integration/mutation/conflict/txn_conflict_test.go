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

func TestTxnConflict_UpdateVsDelete(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnConflict_DeleteVsUpdate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
				ExpectedError: "transaction conflict",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnConflict_DeleteVsDelete(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnConflict_CreateVsCreate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
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
			// Both transactions try to create a document that generates the same ID (same content)
			// Wait, if they have same content, do they get same ID?
			// Defradb document IDs are content-addressable dependent (bae-...).
			// If I create {name: "John", age: 27} twice, they should produce same DocID.
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
						},
					},
				},
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_User(input: {name: "John", age: 27}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
