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

package one_to_one_to_one

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestTxnDeletionOfRelatedDocFromPrimarySideForwardDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"_publisherID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			&action.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e06e5f77-ef19-570a-a866-511e12ed423e") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"_publisherID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			&action.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e06e5f77-ef19-570a-a866-511e12ed423e") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
		// LevelDB does not support concurrent transactions
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"_publisherID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			&action.Request{
				// Delete a linked book that exists.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e06e5f77-ef19-570a-a866-511e12ed423e") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
						},
					},
				},
			},
			&action.Request{
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
								"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
								"name":   "Book By Website",
							},
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
		// LevelDB does not support concurrent transactions
		// TODO https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
				Doc: `{
					"name": "Book By Website",
					"rating": 4.0,
					"_publisherID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85"
				}`,
			},
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
				Doc: `{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
			&action.Request{
				// Delete a linked book that exists in transaction 0.
				TransactionID: immutable.Some(0),
				Request: `mutation {
			        delete_Book(docID: "bae-e06e5f77-ef19-570a-a866-511e12ed423e") {
			            _docID
			        }
			    }`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
						},
					},
				},
			},
			&action.Request{
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
							"_docID": "bae-e06e5f77-ef19-570a-a866-511e12ed423e",
							"name":   "Book By Website",
							"publisher": map[string]any{
								"_docID": "bae-0cd9a444-adb8-59c5-85e1-f95311ee9f85",
								"name":   "Website",
							},
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-2bc16473-47d5-5458-9099-c09ef0361303",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"_publisherID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91"
				}`,
			},
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			&action.Request{
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
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
			&action.AddDoc{
				// books
				CollectionID: 0,
				// "_docID": "bae-2bc16473-47d5-5458-9099-c09ef0361303",
				Doc: `{
					"name": "Book By Online",
					"rating": 4.0,
					"_publisherID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91"
				}`,
			},
			&action.AddDoc{
				// publishers
				CollectionID: 2,
				// "_docID": "bae-0c752d75-5819-599f-ba18-31ee6f177d91",
				Doc: `{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
			&action.Request{
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
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.Request{
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
							"_docID":    "bae-2bc16473-47d5-5458-9099-c09ef0361303",
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
