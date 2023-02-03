// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package relation_create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	relationTests "github.com/sourcenetwork/defradb/tests/integration/mutation/relation"
)

func TestTransactionalCreationAndLinkingOfRelationalDocumentsForward(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Create relational documents, and check the links in forward direction.",

		Docs: map[int][]string{
			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,

				// "_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				`{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
		},

		TransactionalRequests: []testUtils.TransactionRequest{
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			{
				TransactionId: 0,

				Request: `mutation {
					create_book(data: "{\"name\": \"Book By Website\",\"rating\": 4.0, \"publisher_id\": \"bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4\"}") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},

			{
				TransactionId: 1,

				Request: `mutation {
					create_book(data: "{\"name\": \"Book By Online\",\"rating\": 4.0, \"publisher_id\": \"bae-8a381044-9206-51e7-8bc8-dc683d5f2523\"}") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
					},
				},
			},

			// Assert publisher -> books direction within transaction 0.
			{
				TransactionId: 0,

				Request: `query {
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

					{
						"_key":      "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
						"name":      "Online",
						"published": nil,
					},
				},
			},

			// Assert publisher -> books direction within transaction 1.
			{
				TransactionId: 1,

				Request: `query {
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

					{
						"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
						"name": "Online",
						"published": map[string]any{
							"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
							"name": "Book By Online",
						},
					},
				},
			},
		},

		// Assert books -> publisher direction outside the transactions.
		Request: `query {
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

			{
				"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
				"name": "Book By Online",
				"publisher": map[string]any{
					"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
					"name": "Online",
				},
			},
		},
	}

	relationTests.ExecuteTestCase(t, test)
}

func TestTransactionalCreationAndLinkingOfRelationalDocumentsBackward(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Create relational documents, and check the links in backward direction.",

		Docs: map[int][]string{
			// publishers
			2: {
				// "_key": "bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4",
				`{
					"name": "Website",
					"address": "Manning Publications"
				}`,

				// "_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				`{
					"name": "Online",
					"address": "Manning Early Access Program (MEAP)"
				}`,
			},
		},

		TransactionalRequests: []testUtils.TransactionRequest{
			// Create books related to publishers, and ensure they are correctly linked (in and out of transactions).
			{
				TransactionId: 0,

				Request: `mutation {
					create_book(data: "{\"name\": \"Book By Website\",\"rating\": 4.0, \"publisher_id\": \"bae-0e7c3bb5-4917-5d98-9fcf-b9db369ea6e4\"}") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-5b16ccd7-9cae-5145-a56c-03cfe7787722",
					},
				},
			},

			{
				TransactionId: 1,

				Request: `mutation {
					create_book(data: "{\"name\": \"Book By Online\",\"rating\": 4.0, \"publisher_id\": \"bae-8a381044-9206-51e7-8bc8-dc683d5f2523\"}") {
						_key
					}
				}`,

				Results: []map[string]any{
					{
						"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
					},
				},
			},

			// Assert publisher -> books direction within transaction 0.
			{
				TransactionId: 0,

				Request: `query {
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

			// Assert publisher -> books direction within transaction 1.
			{
				TransactionId: 1,

				Request: `query {
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
						"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
						"name": "Book By Online",
						"publisher": map[string]any{
							"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
							"name": "Online",
						},
					},
				},
			},
		},

		// Assert publishers -> books direction outside the transactions.
		Request: `query {
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

			{
				"_key": "bae-8a381044-9206-51e7-8bc8-dc683d5f2523",
				"name": "Online",
				"published": map[string]any{
					"_key": "bae-edf7f0fc-f0fd-57e2-b695-569d87e1b251",
					"name": "Book By Online",
				},
			},
		},
	}

	relationTests.ExecuteTestCase(t, test)
}
