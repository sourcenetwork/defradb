// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
