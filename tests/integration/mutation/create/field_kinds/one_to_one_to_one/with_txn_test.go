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
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-0c752d75-5819-599f-ba18-31ee6f177d91"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
						},
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
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"_docID":    "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
							"name":      "Online",
							"published": nil,
						},
						{
							"_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
								"name":   "Book By Website",
							},
						},
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
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
								"name":   "Book By Online",
							},
						},
						{
							"_docID":    "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
							"name":      "Website",
							"published": nil,
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
								"name":   "Online",
							},
						},
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
								"name":   "Website",
							},
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
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-0c752d75-5819-599f-ba18-31ee6f177d91"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
						},
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
								"name":   "Website",
							},
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
								"name":   "Online",
							},
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
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
								"name":   "Book By Online",
							},
						},
						{
							"_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
								"name":   "Book By Website",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	execute(t, test)
}
