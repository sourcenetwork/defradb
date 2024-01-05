// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestTransactionalCreationAndLinkingOfRelationalDocumentsForward(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create relational documents, and check the links in forward direction.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"}) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"}) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
					},
				},
			},
			// Assert publisher -> books direction within transaction 0.
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					Publisher {
						_docID
						name
						published {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":   "Website",
						"published": map[string]any{
							"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
							"name":   "Book By Website",
						},
					},

					{
						"_docID":    "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
						"name":      "Online",
						"published": nil,
					},
				},
			},
			// Assert publisher -> books direction within transaction 1.
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `query {
					Publisher {
						_docID
						name
						published {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID":    "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":      "Website",
						"published": nil,
					},

					{
						"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
						"name":   "Online",
						"published": map[string]any{
							"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
							"name":   "Book By Online",
						},
					},
				},
			},
			// Commit the transactions before querying the end result
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			testUtils.Request{
				// Assert books -> publisher direction outside the transactions.
				Request: `query {
					Book {
						_docID
						name
						publisher {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
						"name":   "Book By Website",
						"publisher": map[string]any{
							"_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
							"name":   "Website",
						},
					},

					{
						"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
						"name":   "Book By Online",
						"publisher": map[string]any{
							"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
							"name":   "Online",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestTransactionalCreationAndLinkingOfRelationalDocumentsBackward(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create relational documents, and check the links in backward direction.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"}) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"}) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
					},
				},
			},
			// Assert publisher -> books direction within transaction 0.
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					Book {
						_docID
						name
						publisher {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
						"name":   "Book By Website",
						"publisher": map[string]any{
							"_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
							"name":   "Website",
						},
					},
				},
			},
			// Assert publisher -> books direction within transaction 1.
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `query {
					Book {
						_docID
						name
						publisher {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
						"name":   "Book By Online",
						"publisher": map[string]any{
							"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
							"name":   "Online",
						},
					},
				},
			},
			// Commit the transactions before querying the end result
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			testUtils.Request{
				// Assert publishers -> books direction outside the transactions.
				Request: `query {
					Publisher {
						_docID
						name
						published {
							_docID
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":   "Website",
						"published": map[string]any{
							"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
							"name":   "Book By Website",
						},
					},

					{
						"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
						"name":   "Online",
						"published": map[string]any{
							"_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
							"name":   "Book By Online",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
