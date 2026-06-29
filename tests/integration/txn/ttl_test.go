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

package txn_testing

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestTxnTTL_GivenIdleHTTPTransaction_DiscardsTransaction(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        300 * time.Millisecond,
			TxnTTLTick:    50 * time.Millisecond,
			TxnTTLBuckets: 10,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 28}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
			&action.Wait{
				Duration: immutable.Some(900 * time.Millisecond),
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 29}) {
						name
						age
					}
				}`,
				ExpectedError: "transaction not found",
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnTTL_GivenHTTPTransactionActivity_ExtendsTTL(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        500 * time.Millisecond,
			TxnTTLTick:    50 * time.Millisecond,
			TxnTTLBuckets: 20,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 28}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
			&action.Wait{
				Duration: immutable.Some(300 * time.Millisecond),
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 29}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(29),
						},
					},
				},
			},
			&action.Wait{
				Duration: immutable.Some(300 * time.Millisecond),
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(29),
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(29),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnTTL_GivenCommittedHTTPTransaction_RemovesCacheEntry(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        500 * time.Millisecond,
			TxnTTLTick:    50 * time.Millisecond,
			TxnTTLBuckets: 20,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 28}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				ExpectedError: "transaction not found",
			},
			&action.Wait{
				Duration: immutable.Some(600 * time.Millisecond),
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnTTL_GivenDiscardedHTTPTransaction_RemovesCacheEntry(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        500 * time.Millisecond,
			TxnTTLTick:    50 * time.Millisecond,
			TxnTTLBuckets: 20,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 28}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
			&action.DiscardTransaction{
				TransactionID: 1,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				ExpectedError: "transaction not found",
			},
			&action.Wait{
				Duration: immutable.Some(600 * time.Millisecond),
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTxnTTL_GivenExpiredHTTPTransaction_OnGetCollections_ReturnsNonGraphQLError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        300 * time.Millisecond,
			TxnTTLTick:    50 * time.Millisecond,
			TxnTTLBuckets: 20,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.GetCollections{
				TransactionID: immutable.Some(1),
				FilterOptions: options.GetCollections(),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
			&action.Wait{
				Duration: immutable.Some(900 * time.Millisecond),
			},
			&action.GetCollections{
				TransactionID: immutable.Some(1),
				FilterOptions: options.GetCollections(),
				ExpectedError: "transaction not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
