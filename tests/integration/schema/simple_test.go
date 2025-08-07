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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaSimpleCreatesSchemaGivenEmptyType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name:    request.DocIDFieldName,
								Kind:    client.FieldKind_DocID,
								FieldID: "bafyreie6fnppc6bkpo5tifamx3rotptp6mveyz5mvkqldkrojpu5ayds74",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleErrorsGivenDuplicateSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SetupComplete{},
			&action.AddSchema{
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
			&action.AddSchema{
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

func TestSchemaSimpleErrorsGivenDuplicateSchemaInSameSDLMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
					type Users {}
					type Users {}
				`,
				ExpectedError: "collection already exists. Name: Users\ncollection already exists. Name: Users",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenNewTypes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
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
			&action.AddSchema{
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
			&action.AddSchema{
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

func TestSchemaSimpleErrorsGivenTypeWithInvalidFieldTypeMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: NotAType
						age: NotAType
					}
				`,
				ExpectedError: "no type found for given name. Field: age, Kind: NotAType\nno type found for given name. Field: name, Kind: NotAType",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaSimpleCreatesSchemaGivenTypeWithStringField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddSchema{
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
			&action.AddSchema{
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
			&action.AddSchema{
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
			&action.AddSchema{
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

func TestSchemaSimple_WithFloat32Field_CreatesSchemaGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						data: Float32
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
									"name": "Float32",
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

func TestSchemaSimple_WithFloat64Field_CreatesSchemaGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						data: Float64
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
									"name": "Float64",
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

func TestSchemaSimple_WithFloatField_CreatesSchemaGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						data: Float
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
									"name": "Float64",
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

// This test helps to ensure we cover all supported types.
//
// It also documents a bug with graphql-go introspection.
// TODO: https://github.com/sourcenetwork/defradb/issues/3429
func TestSchemaSimple_WithAllTypes_CreatesSchemaGivenTypes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						tBool: Boolean
						tNBoolA: [Boolean]
						tBoolA: [Boolean!]
						tInt: Int
						tNIntA: [Int]
						tIntA: [Int!]
						tDateTime: DateTime
						tFloat: Float
						tNFloatA: [Float]
						tFloatA: [Float!]
						tFloat64: Float64
						tNFloat64A: [Float64]
						tFloat64A: [Float64!]
						tFloat32: Float32
						tNFloat32A: [Float32]
						tFloat32A: [Float32!]
						tString: String
						tNStringA: [String]
						tStringA: [String!]
						tBlob: Blob
						tJSON: JSON
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
								"name": "tBlob",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Blob"},
							},
						).Append(
							Field{
								"name": "tBool",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Boolean"},
							},
						).Append(
							Field{
								"name": "tBoolA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tDateTime",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "DateTime"},
							},
						).Append(
							Field{
								"name": "tFloat",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Float64"},
							},
						).Append(
							Field{
								"name": "tFloat32",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Float32"},
							},
						).Append(
							Field{
								"name": "tFloat32A",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tFloat64",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Float64"},
							},
						).Append(
							Field{
								"name": "tFloat64A",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tFloatA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tInt",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Int"},
							},
						).Append(
							Field{
								"name": "tIntA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tJSON",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "JSON"},
							},
						).Append(
							Field{
								"name": "tNBoolA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tNFloat32A",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tNFloat64A",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tNFloatA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tNIntA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tNStringA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Append(
							Field{
								"name": "tString",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String"},
							},
						).Append(
							Field{
								"name": "tStringA",
								"type": map[string]any{
									"kind": "LIST",
									"name": any(nil)},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
