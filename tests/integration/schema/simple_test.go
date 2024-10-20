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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaSimpleCreatesSchemaGivenEmptyType(t *testing.T) {
	schemaVersionID := "bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
				ExpectedResults: []client.CollectionDescription{
					{
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
							},
						},
					},
				},
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
			testUtils.GetSchema{
				VersionID: immutable.Some(schemaVersionID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersionID,
						Root:      schemaVersionID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
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
				ExpectedError: "collection already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchemaInSameSDL(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
					type Users {}
				`,
				ExpectedError: "collection already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
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

	testUtils.ExecuteTestCase(t, test)
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

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenTypeWithInvalidFieldType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: NotAType
					}
				`,
				ExpectedError: "no type found for given name. Field: name, Kind: NotAType",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
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

	testUtils.ExecuteTestCase(t, test)
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

	testUtils.ExecuteTestCase(t, test)
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

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenTypeWithBlobField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						data: Blob
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
								"name": "data",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Blob",
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

func TestSchemaSimple_WithJSONField_CreatesSchemaGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						data: JSON
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
								"name": "data",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "JSON",
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
