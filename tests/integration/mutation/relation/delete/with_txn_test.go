// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package relation_delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	relationTests "github.com/sourcenetwork/defradb/tests/integration/mutation/relation"
)

func TestTxnDeletionOfRelatedDocFromPrimarySideForwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Delete related doc with transaction from primary side (forward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
				`{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a liniked book that exists.
			{
				TransactionId: 0,

				Query: `mutation {
			        delete_book(id: "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722") {
			            _key
			        }
			    }`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},
		},

		// Assert after transaction(s) have been commited, to ensure the book was deleted.
		Query: `query {
			publisher {
				_key
				name
				published {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{
			{
				"_key":      "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				"name":      "Website",
				"published": nil,
			},
		},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestTxnDeletionOfRelatedDocFromPrimarySideBackwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Delete related doc with transaction from primary side (backward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
				`{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a liniked book that exists.
			{
				TransactionId: 0,

				Query: `mutation {
			        delete_book(id: "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722") {
			            _key
			        }
			    }`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},
		},

		// Assert after transaction(s) have been commited, to ensure the book was deleted.
		Query: `query {
			book {
				_key
				name
				publisher {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnForwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Transaction can read a record that was deleted in a non-commited transaction (forward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
				`{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a liniked book that exists in transaction 0.
			{
				TransactionId: 0,

				Query: `mutation {
			        delete_book(id: "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722") {
			            _key
			        }
			    }`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},

			// Read the book (forward) that was deleted (in the non-commited transaction) in another transaction.
			{
				TransactionId: 1,

				Query: `query {
					publisher {
						_key
						name
						published {
							_key
							name
						}
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
						"name": "Website",
						"published": map[string]any{
							"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
							"name": "Book By Website",
						},
					},
				},
			},
		},

		// Assert after transaction(s) have been commited, to ensure the book was deleted.
		Query: `query {
			publisher {
				_key
				name
				published {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{
			{
				"_key":      "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				"name":      "Website",
				"published": nil,
			},
		},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnBackwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Transaction can read a record that was deleted in a non-commited transaction (backward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
				`{
					"name": "Book By Website",
					"rating": 4.0,
					"publisher_id": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a liniked book that exists in transaction 0.
			{
				TransactionId: 0,

				Query: `mutation {
					delete_book(id: "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722") {
						_key
					}
			    }`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},

			// Read the book (backwards) that was deleted (in the non-commited transaction) in another transaction.
			{
				TransactionId: 1,

				Query: `query {
					book {
						_key
						name
						publisher {
							_key
							name
						}
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
						"name": "Book By Website",
						"publisher": map[string]any{
							"_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
							"name": "Website",
						},
					},
				},
			},
		},

		// Assert after transaction(s) have been commited, to ensure the book was deleted.
		Query: `query {
			book {
				_key
				name
				publisher {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestTxnDeletionOfRelatedDocFromNonPrimarySideForwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Delete related doc with transaction from non-primary side (forward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
				`{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				`{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a publisher and outside the transaction ensure it's linked
			// book gets correctly unlinked too.
			{
				TransactionId: 0,

				Query: `mutation {
					delete_publisher(id: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
					},
				},
			},
		},

		// Assert after transaction(s) have been commited.
		Query: `query {
			publisher {
				_key
				name
				published {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestTxnDeletionOfRelatedDocFromNonPrimarySideBackwardDirection(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Delete related doc with transaction from non-primary side (backward).",

		Docs: map[int][]string{
			// books
			0: {
				// "_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
				`{
					"name": "Book By Online",
					"rating": 4.0,
					"publisher_id": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523"
				}`,
			},

			// publishers
			2: {
				// "_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				`{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
		},

		TransactionalQueries: []testUtils.TransactionQuery{
			// Delete a publisher and outside the transaction ensure it's linked
			// book gets correctly unlinked too.
			{
				TransactionId: 0,

				Query: `mutation {
					delete_publisher(id: "bae-8a381044-9206-51e7-8bc8-dc683d5f2523") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
					},
				},
			},
		},

		// Assert after transaction(s) have been commited.
		Query: `query {
			book {
				_key
				name
				publisher {
					_key
					name
				}
			}
		}`,

		Results: []map[string]any{
			{
				"_key":      "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
				"name":      "Book By Online",
				"publisher": nil,
			},
		},
	}

	relationTests.ExecuteTestCase(t, test)
}
