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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionSimpleAddsColGivenEmptyType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding an empty collection type creates the collection with only the _docID field and exposes it in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
								FieldID: "bafyreihqzhiz3iwro4jozp6kphq4sosg6ccoqcbiaf7rg5dmvea7aux55a",
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

func TestCollectionVersionSimpleErrorsGivenDuplicateCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with the same name as an already-registered collection returns a collection-already-exists error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			testUtils.SetupComplete{},
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
				ExpectedError: "collection already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionSimpleErrorsGivenDuplicateCollectionInSameSDL(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Providing an SDL with two identically named types in the same AddCollection call returns a collection-already-exists error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
					type Users {}
				`,
				ExpectedError: "collection already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionSimpleErrorsGivenDuplicateCollectionInSameSDLMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Providing an SDL with three identically named types returns aggregated collection-already-exists errors for each duplicate.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleAddsCollectionGivenNewTypes(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding multiple distinct collection types in separate calls makes all types accessible via GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleAddsCollectionWithDefaultFieldsGivenEmptyType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "An empty collection type exposes only the default system fields in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleErrorsGivenTypeWithInvalidFieldType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a field of an unrecognised type name returns a no-type-found error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleErrorsGivenTypeWithInvalidFieldTypeMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with multiple fields of unrecognised types returns aggregated no-type-found errors for each field.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleAddsCollectionGivenTypeWithStringField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a String field exposes that field as a SCALAR String in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleErrorsGivenNonNullField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a non-null scalar field (String!) returns an unsupported-non-null-field error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleErrorsGivenNonNullManyRelationField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a non-null element type in a list relation field returns an unsupported-non-null-variant error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimpleAddsCollectionGivenTypeWithBlobField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a Blob field exposes that field as a SCALAR Blob in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimple_WithJSONField_AddsCollectionGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a JSON field exposes that field as a SCALAR JSON in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimple_WithFloat32Field_AddsCollectionGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a Float32 field exposes that field as a SCALAR Float32 in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimple_WithFloat64Field_AddsCollectionGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a Float64 field exposes that field as a SCALAR Float64 in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersionSimple_WithFloatField_AddsCollectionGivenType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with a Float field exposes that field as a SCALAR Float64 (the canonical float type) in GraphQL introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
func TestCollectionVersionSimple_WithAllTypes_AddsCollectionGivenTypes(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a collection with fields of every supported scalar and array type correctly exposes each field with the expected GraphQL kind and name in introspection.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
