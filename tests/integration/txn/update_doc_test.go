// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs UpdateDoc inside of a transaction, and illustrates that committing the transaction
// results in the document being updated in the database.
func TestTxn_UpdateDoc_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
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
			testUtils.UpdateDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"age": 28
				}`,
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs UpdateDoc inside of a transaction, and illustrates that not committing the transaction
// results in the document not yet being updated in the database.
func TestTxn_UpdateDoc_WithoutCommit_DoesNotUpdateDocument(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
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
			testUtils.UpdateDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"age": 28
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs UpdateDoc inside of a transaction, and illustrates that it can work on
// the documents created inside that transaction.
func TestTxn_UpdateDoc_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(1),
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.UpdateDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"age": 28
				}`,
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs UpdateWithFilter inside of a transaction, and illustrates that committing the transaction
// results in the document being updated in the database.
func TestTxn_UpdateDocWithFilter_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
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
			testUtils.UpdateWithFilter{
				TransactionID: immutable.Some(1),
				Filter:        `{name: {_eq: "John"}}`,
				Updater:       `{"name": "Chris"}`,
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Chris",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs UpdateWithFilter inside of a transaction, and illustrates that not committing
// the transaction results in the document not yet being updated in the database.
func TestTxn_UpdateDocWithFilter_WithoutCommit_DoesNotUpdateDocument(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
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
			testUtils.UpdateWithFilter{
				TransactionID: immutable.Some(1),
				Filter:        `{name: {_eq: "John"}}`,
				Updater:       `{"name": "Chris"}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// todo: The following test should be enabled once any bugs inside UpdateWithFilter are resolved
// see: https://github.com/sourcenetwork/defradb/issues/4614

// This test runs UpdateWithFilter inside of a transaction, and illustrates that it can work on
// the documents created inside that transaction.
/*
func TestTxn_UpdateWithFilter_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(1),
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.UpdateWithFilter{
				TransactionID: immutable.Some(1),
				Filter:        `{name: {_eq: "John"}}`,
				Updater:       `{"name": "Chris"}`,
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Chris",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
*/
