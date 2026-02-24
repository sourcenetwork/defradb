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

	// Add a schema, without a transaction
	userSchema := `
		type Users {
			name: String
			age: Int
		}
	`
	_, err = node.DB.AddSchema(ctx, userSchema)
	if err != nil {
		panic(fmt.Sprintf("Error adding schema: %v", err))
	}
	fmt.Println("Added schema.")

	// Create a transaction
	txn, err := node.DB.NewTxn(false)
	if err != nil {
		panic(fmt.Sprintf("Transaction create error: %v", err))
	}
	fmt.Printf("Transaction created.\n")
	fmt.Printf("Transaction ID: %d\n", txn.ID())

	// Get the collection (with transaction)...
	collectionOpts := options.GetCollections()
	collectionOpts.SetCollectionName("Users")
	cols, err := txn.GetCollections(ctx, collectionOpts)
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

	// Commit the transaction. Note that this is commented out.
	// This shouldn't work when it's commented out, but it does.

	/*
		err = txn.Commit()
		if err != nil {
			panic(fmt.Sprintf("Commit error: %v", err))
		} else {
			fmt.Println("Transaction committed.")
		}
	*/

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

func TestTxn_AddSchema_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(0),
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
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

func TestTxn_CreateDoc_Succeeds(t *testing.T) {
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
				//TransactionID: immutable.Some(0),
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
