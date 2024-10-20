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
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-21084f46-b12a-53ab-94dd-04d075b4218c"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
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
							"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
								"name":   "Book By Website",
							},
						},

						{
							"_docID":    "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
							"name":      "Online",
							"published": nil,
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
							"_docID":    "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
							"name":      "Website",
							"published": nil,
						},
						{
							"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
								"name":   "Book By Online",
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
							"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
								"name":   "Website",
							},
						},
						{
							"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
								"name":   "Online",
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
		Description: "Create relational documents, and check the links in backward direction.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-21084f46-b12a-53ab-94dd-04d075b4218c"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
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
							"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
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
							"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
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
							"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-5a378128-1b3f-50e7-a5ff-027e707c4b87",
								"name":   "Book By Website",
							},
						},

						{
							"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-787391fb-86f8-5cbe-8fc2-ad59f90e267a",
								"name":   "Book By Online",
							},
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
