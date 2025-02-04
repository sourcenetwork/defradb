// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package constraints

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithStringForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [String!] @embedding
					}
				`,
				ExpectedError: "invalid type for vector embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithIntForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Int!] @embedding
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithFloatForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float!] @embedding
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithFloat64ForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float64!] @embedding
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithNillableFloat32ForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float32] @embedding
					}
				`,
				ExpectedError: "invalid type for vector embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithFloat32ForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float32!] @embedding
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithNonExistantField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						name_v: [Float32!] @embedding(fields: ["name", "about"])
					}
				`,
				ExpectedError: "the given field does not exist. Embedding field: about",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithInvalidEmbeddingGenerationFieldType_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
						name_v: [Float32!] @embedding(fields: ["name", "custom"])
					}
				`,
				ExpectedError: "invalid field type for vector embedding generation. Actual: JSON",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithMultipleEmbeddingFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with multiple embedding fields",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			testUtils.CreateDoc{
				// Doc with both embedding fields
				Doc: `{
					"name": "John",
					"about": "He loves tacos."
				}`,
			},
			testUtils.CreateDoc{
				// Doc with only one embedding field
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						User {
							_docID
							name_v
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name_v": testUtils.NewArrayDescription[float32](768),
						},
						{
							"name_v": testUtils.NewArrayDescription[float32](768),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_UserDefinedVectorEmbeddingDoesNotTriggerGeneration_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with manually defined vector embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"about": "He loves tacos."
					"name_v": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						User {
							_docID
							name_v
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 0),
							"name_v": []float32{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
