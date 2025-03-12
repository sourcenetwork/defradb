// Copyright 2025 Democratized Data Foundation
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

func TestSchema_WithStringForEmbedding_ShouldError(t *testing.T) {
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

func TestSchema_WithIntForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Int!] @embedding
					}
				`,
				ExpectedError: "invalid type for vector embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithIntForEmbedding_ShouldError_Multiple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Int!] @embedding
					}
				`,
				ExpectedError: "invalid type for vecdddtor embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithFloatForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float!] @embedding
					}
				`,
				ExpectedError: "invalid type for vector embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithFloat64ForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name_v: [Float64!] @embedding
					}
				`,
				ExpectedError: "invalid type for vector embedding",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithNillableFloat32ForEmbedding_ShouldError(t *testing.T) {
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

func TestSchema_WithFloat32ForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with invalid type for embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithNonExistantFieldForEmbedding_ShouldError(t *testing.T) {
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
				ExpectedError: "the given field does not exist. Embedding generation field: about",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithInvalidEmbeddingGenerationFieldType_ShouldError(t *testing.T) {
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

func TestSchema_WithUnsupportedProviderForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "foo")
					}
				`,
				ExpectedError: "unknown embedding provider",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithMissingModelForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "ollama")
					}
				`,
				ExpectedError: "embedding Model cannot be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_ReferenceToSelfForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name_v"], provider: "ollama", model: "nomic-embed-text")
					}
				`,
				ExpectedError: "embedding fields cannot refer to self or another embedding field. Field: name_v",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_ReferenceToAnotherEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["about_v"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
						about_v: [Float32!] @embedding(fields: ["about"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
				ExpectedError: "embedding fields cannot refer to self or another embedding field. Field: about_v",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
