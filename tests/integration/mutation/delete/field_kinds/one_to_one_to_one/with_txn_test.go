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

func TestTxnDeletionOfRelatedDocFromPrimarySideForwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited, to ensure the book was deleted.
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
							"_docID":    "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
							"name":      "Website",
							"published": nil,
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestTxnDeletionOfRelatedDocFromPrimarySideBackwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited, to ensure the book was deleted.
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
					"Book": []map[string]any{},
				},
			},
		},
	}

	execute(t, test)
}

func TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnForwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.Request{
				// Read the book (forward) that was deleted (in the non-commited transaction) in another transaction.
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited, to ensure the book was deleted.
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
							"_docID":    "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
							"name":      "Website",
							"published": nil,
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnBackwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists in transaction 0.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-f412a4b4-1a86-54c2-9523-73e2f66d6e96",
						},
					},
				},
			},
			testUtils.Request{
				// Read the book (backwards) that was deleted (in the non-commited transaction) in another transaction.
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
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited, to ensure the book was deleted.
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
					"Book": []map[string]any{},
				},
			},
		},
	}

	execute(t, test)
}

func TestTxnDeletionOfRelatedDocFromNonPrimarySideForwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-0c752d75-5819-599f-ba18-31ee6f177d91"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			testUtils.Request{
				// Delete a publisher and outside the transaction ensure it's linked
				// book gets correctly unlinked too.
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_Publisher(docID: "bae-0c752d75-5819-599f-ba18-31ee6f177d91") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Publisher": []map[string]any{
						{
							"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited.
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
					"Publisher": []map[string]any{},
				},
			},
		},
	}

	execute(t, test)
}

func TestTxnDeletionOfRelatedDocFromNonPrimarySideBackwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-0c752d75-5819-599f-ba18-31ee6f177d91"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			testUtils.Request{
				// Delete a publisher and outside the transaction ensure it's linked
				// book gets correctly unlinked too.
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_Publisher(docID: "bae-0c752d75-5819-599f-ba18-31ee6f177d91") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Publisher": []map[string]any{
						{
							"_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
						},
					},
				},
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			testUtils.Request{
				// Assert after transaction(s) have been commited.
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
							"_docID":    "bae-cd5d64a6-90ff-5a59-8a40-3d8ffd42752a",
							"name":      "Book By Online",
							"publisher": nil,
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
