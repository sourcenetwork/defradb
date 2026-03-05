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

package upsert

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpsertSimple_WithNoFilterMatch_AddsNewDoc(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Bob", age: 40},
						update: {age: 40}
					) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"upsert_Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Alice",
							"age":  int64(40),
						},
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatch_UpdatesDoc(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Bob", age: 40},
						update: {age: 40}
					) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"upsert_Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
						{
							"name": "Alice",
							"age":  int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatchOnSameField_UpdatesDoc(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Bob", age: 40},
						update: {name: "John"}
					) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"upsert_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(30),
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Alice",
							"age":  int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatchMultiple_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Bob",
					"age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						add: {name: "Alice", age: 40},
						update: {age: 50}
					) {
						name
						age
					}
				}`,
				ExpectedError: `cannot upsert multiple matching documents`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithNullAddInput_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						add: null,
						update: {age: 50}
					) {
						name
						age
					}
				}`,
				ExpectedError: `Argument "add" has invalid value <nil>`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithNullUpdateInput_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						add: {name: "Alice", age: 40},
						update: null,
					) {
						name
						age
					}
				}`,
				ExpectedError: `Argument "update" has invalid value <nil>`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithNullFilterInput_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: null,
						add: {name: "Alice", age: 40},
						update: {age: 50}
					) {
						name
						age
					}
				}`,
				ExpectedError: `Argument "filter" has invalid value <nil>`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithUniqueCompositeIndexAndDuplicateUpdate_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @index(includes: [{field: "name"}, {field: "age"}], unique: true) {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"age": 50
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Alice", age: 40},
						update: {name: "Alice", age: 40}
					) {
						name
						age
					}
				}`,
				ExpectedError: `can not index a doc's field(s) that violates unique index`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatchAndVersion_UpdatesDoc(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Alice",
					"age": 40
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			&action.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Bob", age: 40},
						update: {age: 40}
					) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"upsert_Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
						{
							"name": "Alice",
							"age":  int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
