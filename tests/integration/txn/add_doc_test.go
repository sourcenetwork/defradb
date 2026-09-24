// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// This test runs AddDoc inside of a transaction, and illustrates that committing the transaction
// results in the document being added to the database.
func TestTxn_AddDoc_WithCommit_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
							}
						}
					`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
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

// This test runs AddDoc inside of a transaction, and illustrates that not committing the transaction
// results in the document not yet being in the database.
func TestTxn_AddDoc_WithoutCommit_EmptyResults(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddDoc inside of a transaction, and illustrates that a query running inside
// the same transaction can see the uncommitted document, while a query outside the transaction
// cannot — mirroring the isolation behaviour exercised for UpdateDoc and DeleteDoc in this
// package. See GH issue #4475.
func TestTxn_AddDoc_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			// The query inside the transaction sees the doc that was just added.
			&action.Request{
				TransactionID: immutable.Some(1),
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
			// The query outside the transaction does not see the uncommitted doc.
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
					"Users": []map[string]any{},
				},
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			// After commit, the doc is visible globally.
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

// This test runs multiple AddDoc actions inside the same transaction and illustrates that
// committing the transaction results in all of the documents being added atomically.
// See GH issue #4475.
func TestTxn_AddManyDocs_WithCommit_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "Jane",
					"age": 30
				}`,
			},
			&action.CommitTransaction{
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
				NonOrderedResults: true,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
						{
							"name": "Jane",
							"age":  int64(30),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs multiple AddDoc actions inside the same transaction and illustrates that
// without a commit none of the documents are visible outside the transaction.
// See GH issue #4475.
func TestTxn_AddManyDocs_WithoutCommit_EmptyResults(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "Jane",
					"age": 30
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
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddDoc inside a transaction and then discards the transaction, illustrating
// that the document never becomes visible. See GH issue #4475.
func TestTxn_AddDoc_WithDiscard_NeverVisible(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.DiscardTransaction{
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
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddDoc and UpdateDoc inside the same transaction, illustrating that the
// update can target a document that does not exist outside the transaction, and that on
// commit the document is visible at its post-update state. See GH issue #4475.
func TestTxn_AddDoc_ThenUpdateSameDoc_InSameTxn_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.UpdateDoc{
				TransactionID: immutable.Some(1),
				DocID:         0,
				Doc: `{
					"age": 28
				}`,
			},
			&action.CommitTransaction{
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

// This test runs AddDoc and DeleteDoc inside the same transaction, illustrating that the
// delete cancels the add and on commit no document is visible. See GH issue #4475.
func TestTxn_AddDoc_ThenDeleteSameDoc_InSameTxn_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.DeleteDoc{
				TransactionID: immutable.Some(1),
				CollectionID:  0,
				DocID:         0,
			},
			&action.CommitTransaction{
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
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs two concurrent transactions that both try to add a document with the same
// docID, illustrating the conflict semantics on the second commit. See GH issue #4475.
//
// The expected error string was captured from the actual implementation behaviour; if the
// behaviour ever changes we want this test to flag that explicitly.
func TestTxn_AddDoc_ConcurrentTxnsSameDocID_SecondCommitConflicts(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(2),
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.CommitTransaction{
				TransactionID: 2,
				ExpectedError: "transaction conflict. Please retry",
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

// This test runs AddDoc inside of the same transaction as AddCollection, and illustrates that committing the transaction
// results in the document being added to the database.
func TestTxn_AddDoc_InsideTxnWithAddCollection_WithCommit_Succeeds(t *testing.T) {
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
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
							}
						}
					`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
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
