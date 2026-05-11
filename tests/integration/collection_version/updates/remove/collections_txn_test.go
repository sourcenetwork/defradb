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

package remove

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestColVersionUpdateRemoveCollection_WithDataAddedInSameTxn(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.PatchCollection{
				TransactionID: immutable.Some(0),
				Patch: `
						[
							{
								"op": "remove",
								"path": "/Users"
							}
						]
					`,
				ExpectedError: "cannot delete a collection that has documents",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollection_DeadlocksIfOtherTxnWriting(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			// Badger file fails after the tests successfully completes, probably caused by the
			// leaked database instance.  It is not worth the time to chase down at the moment.
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
			testUtils.LevelStoreType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
		},
	}

	done := make(chan struct{})
	go func() {
		testUtils.ExecuteTestCase(t, test)
		done <- struct{}{}
	}()

	select {
	case <-time.After(100 * time.Millisecond):
		return
	case <-done:
		t.Fatal("Test should deadlock but did not")
	}
}

func TestColVersionUpdateRemoveCollection_DeadlockIfDeletingVersionWithNewFieldWhilstOtherTxnWriting(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			// Badger file fails after the tests successfully completes, probably caused by the
			// leaked database instance.  It is not worth the time to chase down at the moment.
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
			testUtils.LevelStoreType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
						}
					]
				`,
			},
		},
	}

	done := make(chan struct{})
	go func() {
		testUtils.ExecuteTestCase(t, test)
		done <- struct{}{}
	}()

	select {
	case <-time.After(100 * time.Millisecond):
		return
	case <-done:
		t.Fatal("Test should deadlock but did not")
	}
}

func TestColVersionUpdateRemoveCollection_GetCollectionsShouldNotReturnCollectionDeletedWhilstTxnWasOpen(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			// LevelDB is not supported for this test as the test opens multiple transactions at
			// the same time.
			testUtils.BadgerIMType,
			testUtils.BadgerFileType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddCollection{
				// We don't yet have a create transaction action, so we need to create
				// one by performing an operation on anything but the `Users` collection
				// as doing anything on `Users` will acquire a read lock on the collection.
				TransactionID: immutable.Some(0),
				SDL: `
					type Books {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				TransactionID: immutable.Some(0),
				// Users must not be returned, even though it was deleted after this transaction
				// was created.
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollection_CollectionMayBeRedeclaredAndUsedByTxn(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			// LevelDB is not supported for this test as the test opens multiple transactions at
			// the same time.
			testUtils.BadgerIMType,
			testUtils.BadgerFileType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddCollection{
				// We don't yet have a create transaction action, so we need to create
				// one by performing an operation on anything but the `Users` collection
				// as doing anything on `Users` will acquire a read lock on the collection.
				TransactionID: immutable.Some(0),
				SDL: `
					type Books {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.AddCollection{
				TransactionID: immutable.Some(0),
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
