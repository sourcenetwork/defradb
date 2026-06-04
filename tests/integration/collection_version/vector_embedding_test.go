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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersion_WithStringForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithIntForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
func TestCollectionVersion_WithFloatForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithFloat64ForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithNillableFloat32ForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithFloat32ForEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithNonExistantFieldForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithInvalidEmbeddingGenerationFieldType_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithUnsupportedProviderForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_WithMissingModelForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_ReferenceToSelfForEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

func TestCollectionVersion_ReferenceToAnotherEmbedding_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
