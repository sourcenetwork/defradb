// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package upsert

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpsertSimple_WithNoFilterMatch_CreatesNewDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple upsert mutation with no filter match",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						create: {name: "Bob", age: 40},
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
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatch_UpdatesDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple upsert mutation with filter match",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						create: {name: "Bob", age: 40},
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
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithFilterMatchMultiple_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple upsert mutation with multiple filter matches",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						create: {name: "Alice", age: 40},
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

func TestMutationUpsertSimple_WithNullCreateInput_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple upsert mutation with null create input",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						create: null,
						update: {age: 50}
					) {
						name
						age
					}
				}`,
				ExpectedError: `Argument "create" has invalid value <nil>`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpsertSimple_WithNullUpdateInput_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple upsert mutation with null update input",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {},
						create: {name: "Alice", age: 40},
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
		Description: "Simple upsert mutation with null filter input",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: null,
						create: {name: "Alice", age: 40},
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
		Description: "Simple upsert mutation with unique composite index and update",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @index(includes: [{field: "name"}, {field: "age"}], unique: true) {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Alice",
					"age": 40
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"age": 50
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					upsert_Users(
						filter: {name: {_eq: "Bob"}},
						create: {name: "Alice", age: 40},
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
