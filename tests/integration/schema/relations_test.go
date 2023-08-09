// Copyright 2023 Democratized Data Foundation
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

func TestSchemaRelationOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
						user: User
					}
					type User {
						dog: Dog
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "User") {
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
						"name": "User",
						"fields": append(DefaultFields,
							Field{
								"name": "dog",
								"type": map[string]any{
									"kind": "OBJECT",
									"name": "Dog",
								},
							},
							Field{
								"name": "dog_id",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaRelationManyToOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
						user: User
					}
					type User {
						dogs: [Dog]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "User") {
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
						"name": "User",
						"fields": append(DefaultFields,
							Field{
								"name": "dogs",
								"type": map[string]any{
									"kind": "LIST",
									"name": nil,
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaRelationErrorsGivenOneSidedManyRelationField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
					}
					type User {
						dogs: [Dog]
					}
				`,
				ExpectedError: "relation must be defined on both schemas. Field: dogs, Type: Dog",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaRelationErrorsGivenOneSidedRelationField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
					}
					type User {
						dog: Dog
					}
				`,
				ExpectedError: "relation must be defined on both schemas. Field: dog, Type: Dog",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaRelation_GivenSelfReferemceRelationField_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Dog {
						name: String
						bestMate: Dog
					}
				`,
				ExpectedError: "relation must be defined on both schemas. Field: bestMate, Type: Dog",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
