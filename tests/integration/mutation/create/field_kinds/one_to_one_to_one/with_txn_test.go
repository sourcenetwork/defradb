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
				// "_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
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
							"_docID":    "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
							"name":      "Online",
							"published": nil,
						},
						{
							"_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
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
							"_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
								"name":   "Book By Online",
							},
						},
						{
							"_docID":    "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
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
							"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
								"name":   "Online",
							},
						},
						{
							"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
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
		Description: "Create relational documents, and check the links in backward direction.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				// "_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			testUtils.Request{
				TransactionID: immutable.Some(0),
				Request: `mutation {
					create_Book(input: {name: "Book By Website", rating: 4.0, publisher_id: "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
						},
					},
				},
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					create_Book(input: {name: "Book By Online", rating: 4.0, publisher_id: "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81"}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"create_Book": []map[string]any{
						{
							"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
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
							"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
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
							"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
							"name":   "Book By Online",
							"publisher": map[string]any{
								"_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
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
							"_docID": "bae-2cfd54a8-7f18-5354-a308-805ba1d68f81",
							"name":   "Online",
							"published": map[string]any{
								"_docID": "bae-635d5e56-599c-52df-842f-a5a2f0bc6c02",
								"name":   "Book By Online",
							},
						},
						{
							"_docID": "bae-a69fd0a0-eb5b-53f1-aeed-8833da8c9cc5",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-85148ba6-74ea-560a-820d-adf0b4b05531",
								"name":   "Book By Website",
							},
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
