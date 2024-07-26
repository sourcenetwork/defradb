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
		Description: "Delete related doc with transaction from primary side (forward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e7943028-5c74-5fd4-9661-0a233edcd287") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
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
							"_docID":    "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
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
		Description: "Delete related doc with transaction from primary side (backward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e7943028-5c74-5fd4-9661-0a233edcd287") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
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
		Description: "Transaction can read a record that was deleted in a non-commited transaction (forward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e7943028-5c74-5fd4-9661-0a233edcd287") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
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
							"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
							"name":   "Website",
							"published": map[string]any{
								"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
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
							"_docID":    "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
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
		Description: "Transaction can read a record that was deleted in a non-commited transaction (backward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists in transaction 0.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e7943028-5c74-5fd4-9661-0a233edcd287") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
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
							"_docID": "bae-e7943028-5c74-5fd4-9661-0a233edcd287",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-07fd000a-d023-54b9-b8f3-a4318fac8fed",
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
		Description: "Delete related doc with transaction from non-primary side (forward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-7f6a5a76-b90d-5715-a452-708ded9e7ae7",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-21084f46-b12a-53ab-94dd-04d075b4218c"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
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
					delete_Publisher(docID: "bae-21084f46-b12a-53ab-94dd-04d075b4218c") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Publisher": []map[string]any{
						{
							"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
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
		Description: "Delete related doc with transaction from non-primary side (backward).",
		Actions: []any{
			testUtils.CreateDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-7f6a5a76-b90d-5715-a452-708ded9e7ae7",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-21084f46-b12a-53ab-94dd-04d075b4218c"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
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
					delete_Publisher(docID: "bae-21084f46-b12a-53ab-94dd-04d075b4218c") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Publisher": []map[string]any{
						{
							"_docID": "bae-21084f46-b12a-53ab-94dd-04d075b4218c",
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
							"_docID":    "bae-7f6a5a76-b90d-5715-a452-708ded9e7ae7",
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
