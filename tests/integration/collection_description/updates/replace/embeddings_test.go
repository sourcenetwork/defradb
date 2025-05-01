// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdate_ReplaceVectorEmbeddingWithUnknownFieldName_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/FieldName",
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

func TestColDescrUpdate_ReplaceVectorEmbeddingWithUnknownEmbeddingGenerationField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/Fields",
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

func TestColDescrUpdate_ReplaceVectorEmbeddingWithInvalidEmbeddingGenerationFieldKind_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreidjpk55vtz3l5ouzkgfbxv2rnt3xekjb5aide7i246ngsqtgcbroa/VectorEmbeddings/0/Fields",
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

func TestColDescrUpdate_ReplaceVectorEmbeddingParams_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/Fields",
							"value": ["about"]
						},
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/Provider",
							"value": "ollama"
						},
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/Model",
							"value": "nomic-embed-text"
						},
						{
							"op": "replace",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0/URL",
							"value": "http://localhost:11434/api"
						}
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
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
