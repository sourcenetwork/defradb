// Copyright 2025 Democratized Data Foundation
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
	"context"
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestTxnBehavior(t *testing.T) {
	// Create a context to use throughout
	ctx := context.Background()

	// Create and start the node
	opts := options.Node().SetEnableDevelopment(false).SetDisableP2P(false)
	opts.Store().SetBadgerInMemory(true)
	node, err := node.New(ctx, opts)
	if err != nil {
		panic(fmt.Sprintf("Node error: %v", err))
	}
	fmt.Println("Node created.")

	err = node.Start(ctx)
	if err != nil {
		panic(fmt.Sprintf("Error starting node: %v", err))
	}
	fmt.Println("Node started.")

	defer node.Close(ctx)

	// Create a transaction
	txn, err := node.DB.NewTxn(false)
	if err != nil {
		panic(fmt.Sprintf("Transaction create error: %v", err))
	}
	fmt.Printf("Transaction created.\n")
	fmt.Printf("Transaction ID: %d\n", txn.ID())

	// Add a schema, with a transaction
	userSchema := `
		type Users {
			name: String
			age: Int
		}
	`
	_, err = txn.AddSchema(ctx, userSchema)
	if err != nil {
		panic(fmt.Sprintf("Error adding schema: %v", err))
	}

	// Commit the transaction.
	err = txn.Commit()
	if err != nil {
		panic(fmt.Sprintf("Commit error: %v", err))
	} else {
		fmt.Println("Transaction committed.")
	}

	fmt.Println("Added schema.")

	// Create a transaction
	txn2, err := node.DB.NewTxn(false)
	if err != nil {
		panic(fmt.Sprintf("Transaction create error: %v", err))
	}
	fmt.Printf("Transaction created.\n")
	fmt.Printf("Transaction ID: %d\n", txn2.ID())

	// Get the collection (with transaction)...
	collectionOpts := options.GetCollections()
	collectionOpts.SetCollectionName("Users")
	cols, err := txn2.GetCollections(ctx, collectionOpts)
	if err != nil {
		panic(fmt.Sprintf("Error getting collections: %v", err))
	}
	if len(cols) != 1 {
		panic(fmt.Sprintf("Expected 1 collection, got %d", len(cols)))
	}
	collection := cols[0]

	// ...and create some documents (with transaction)
	docJSON := `[{"name":"Alica","age":16},{"name":"Verso","age":26}]`
	doc, err := client.NewDocsFromJSON(ctx, []byte(docJSON), collection.Version())
	if err != nil {
		panic(fmt.Sprintf("Error creating document: %v", err))
	}

	err = collection.AddMany(ctx, doc)
	if err != nil {
		panic(fmt.Sprintf("Error creating document: %v", err))
	}

	fmt.Println("Created documents.")

	// Commit the transaction.

	err = txn2.Commit()
	if err != nil {
		panic(fmt.Sprintf("Commit error: %v", err))
	} else {
		fmt.Println("Transaction committed.")
	}

	// Make a query for the documents
	query := `
		query {
			Users {
				name
				age
			}
		}
	`

	freshCtx := context.Background()
	res := node.DB.ExecRequest(freshCtx, query)

	fmt.Println("Read Result:")
	fmt.Println(res)
}

// This test runs AddSchema inside of a transaction, and illustrates that committing the transaction
// results in the schema being added to the database.
func TestTxn_AddSchema_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			&action.AddDoc{
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
					"Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
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

// This test runs AddSchema inside of a transaction, and illustrates that not committing the transaction
// results in the schema not being ready for use.
func TestTxn_AddSchema_Fails(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
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
				ExpectedError: "key not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddDoc inside of a transaction, and illustrates that committing the transaction
// results in the document being added to the database.
func TestTxn_AddDoc_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			testUtils.TransactionCommit{
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
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
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
func TestTxn_AddDoc_WithCommit_EmptyResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
