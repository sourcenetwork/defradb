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

package delete

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestDeleteWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithMapFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter: map[string]any{
					"name": map[string]any{"_eq": "John"},
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithMapFilter_LargeIntegerStraddling2To53_DeletesOnlyExactMatch(t *testing.T) {
	test := testUtils.TestCase{
		// JS numbers are IEEE-754 doubles, so an integer filter condition above 2^53 has
		// already lost precision before it reaches Go. Unlike the other clients.
		// To do: https://github.com/sourcenetwork/defradb/issues/5176.
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 9007199254740992
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 9007199254740993
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter: map[string]any{
					"age": map[string]any{"_eq": int64(9007199254740993)},
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithSomeOptionFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter: immutable.Some(request.Filter{
					Conditions: map[string]any{
						"name": map[string]any{"_eq": "John"},
					},
				}),
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithNoneOptionFilter_DeletesAllDocuments(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       immutable.None[request.Filter](),
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteWithFilter_WithMultipleMatchingDocs_DeletesAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.DeleteWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
