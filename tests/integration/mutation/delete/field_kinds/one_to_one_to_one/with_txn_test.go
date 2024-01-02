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
				// "_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(id: "bae-37de3681-1856-5bc9-9fd6-1595647b7d96") {
			            _docID
			        }
			    }`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
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
				Results: []map[string]any{
					{
						"_docID":    "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":      "Website",
						"published": nil,
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
				// "_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(id: "bae-37de3681-1856-5bc9-9fd6-1595647b7d96") {
			            _docID
			        }
			    }`,
				Results: []map[string]any{
					{
						"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
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
				Results: []map[string]any{},
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
				// "_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(id: "bae-37de3681-1856-5bc9-9fd6-1595647b7d96") {
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
				Results: []map[string]any{
					{
						"_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":   "Website",
						"published": map[string]any{
							"_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
							"name":   "Book By Website",
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
				Results: []map[string]any{
					{
						"_docID":    "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name":      "Website",
						"published": nil,
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
				// "_docID": "bae-37de3681-1856-5bc9-9fd6-1595647b7d96",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			testUtils.Request{
				// Delete a linked book that exists in transaction 0.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(id: "bae-37de3681-1856-5bc9-9fd6-1595647b7d96") {
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
				Results: []map[string]any{},
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
				// "_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
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
					delete_Publisher(docID: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523") {
			            _docID
			        }
			    }`,
				Results: []map[string]any{
					{
						"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
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
				Results: []map[string]any{},
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
				// "_docID": "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"
				}`,
			},
			testUtils.CreateDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
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
					delete_Publisher(docID: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523") {
			            _docID
			        }
			    }`,
				Results: []map[string]any{
					{
						"_docID": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
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
				Results: []map[string]any{
					{
						"_docID":    "bae-60ffc9b4-0e31-5d63-82dc-c5cb007f2985",
						"name":      "Book By Online",
						"publisher": nil,
					},
				},
			},
		},
	}

	execute(t, test)
}
