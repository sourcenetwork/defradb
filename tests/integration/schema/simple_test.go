// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaSimpleCreatesSchemaGivenEmptyType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SetupComplete{},
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
				ExpectedError: "schema type already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchemaInSameSDL(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
					type Users {}
				`,
				ExpectedError: "schema type already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleCreatesSchemaGivenNewTypes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Books") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Books",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users", "Books"}, test)
}

func TestSchemaSimpleCreatesSchemaWithDefaultFieldsGivenEmptyType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name":   "Users",
						"fields": DefaultFields.Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleCreatesSchemaGivenTypeWithStringField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
						"fields": DefaultFields.Append(
							Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleErrorsGivenNonNullField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						email: String!
					}
				`,
				ExpectedError: "NonNull fields are not currently supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaSimpleErrorsGivenNonNullManyRelationField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dogs {
						name: String
						user: Users
					}
					type Users {
						Dogs: [Dogs!]
					}
				`,
				ExpectedError: "NonNull variants for type are not supported. Type: Dogs",
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Dogs", "Users"}, test)
}
