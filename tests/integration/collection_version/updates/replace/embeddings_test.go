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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdate_ReplaceVectorEmbeddingWithUnknownFieldName_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/FieldName",
							"value": "foo"
						}
					]
				`,
				ExpectedError: "the given field does not exist. Vector field: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdate_ReplaceVectorEmbeddingWithUnknownEmbeddingGenerationField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/Fields",
							"value": ["name", "foo"]
						}
					]
				`,
				ExpectedError: "the given field does not exist. Embedding generation field: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdate_ReplaceVectorEmbeddingWithInvalidEmbeddingGenerationFieldKind_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						custom: JSON
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/Fields",
							"value": ["name", "custom"]
						}
					]
				`,
				ExpectedError: "invalid field type for vector embedding generation. Actual: JSON",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdate_ReplaceVectorEmbeddingParams_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/Fields",
							"value": ["about"]
						},
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/Provider",
							"value": "ollama"
						},
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/Model",
							"value": "nomic-embed-text"
						},
						{
							"op": "replace",
							"path": "/Users/VectorEmbeddings/0/URL",
							"value": "http://localhost:11434/api"
						}
					]
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						VectorEmbeddings: []client.VectorEmbeddingDescription{
							{
								FieldName: "name_v",
								Fields:    []string{"about"},
								Provider:  "ollama",
								Model:     "nomic-embed-text",
								URL:       "http://localhost:11434/api",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
