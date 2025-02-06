// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdate_AddVectorEmbeddingWithUnknownFieldName_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "foo"} }
					]
				`,
				ExpectedError: "the given field does not exist. Vector field: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithUnknownEmbeddingGenerationField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "foo"]} }
					]
				`,
				ExpectedError: "the given field does not exist. Embedding generation field: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithInvalidEmbeddingGenerationFieldKind_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "custom"]} }
					]
				`,
				ExpectedError: "invalid field type for vector embedding generation. Actual: JSON",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "about"], "Provider": "ollama", "Model": "nomic-embed-text",  "URL": "http://localhost:11434/api"} }
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						Embeddings: []client.EmbeddingDescription{
							{
								FieldName: "name_v",
								Fields:    []string{"name", "about"},
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

func TestColDescrUpdate_AddVectorEmbeddingWithMissingFieldName_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"Fields": ["name", "about"], "Provider": "ollama", "Model": "nomic-embed-text",  "URL": "http://localhost:11434/api"} }
					]
				`,
				ExpectedError: "embedding FieldName cannot be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithMissingFields_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Provider": "ollama", "Model": "nomic-embed-text",  "URL": "http://localhost:11434/api"} }
					]
				`,
				ExpectedError: "embedding Fields cannot be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithMissingProvider_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "about"], "Model": "nomic-embed-text",  "URL": "http://localhost:11434/api"} }
					]
				`,
				ExpectedError: "embedding Provider cannot be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithUnsupportedgProvider_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "about"], "Provider": "deepseek", "Model": "nomic-embed-text",  "URL": "http://localhost:11434/api"} }
					]
				`,
				ExpectedError: "unknown embedding provider",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithMissingModel_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "about"], "Provider": "ollama",  "URL": "http://localhost:11434/api"} }
					]
				`,
				ExpectedError: "embedding Model cannot be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdate_AddVectorEmbeddingWithMissingURL_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!]
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/1/Embeddings/-", "value": {"FieldName": "name_v", "Fields": ["name", "about"], "Provider": "ollama", "Model": "nomic-embed-text"} }
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name:           immutable.Some("Users"),
						IsMaterialized: true,
						Embeddings: []client.EmbeddingDescription{
							{
								FieldName: "name_v",
								Fields:    []string{"name", "about"},
								Provider:  "ollama",
								Model:     "nomic-embed-text",
								// URL is not a required parameted. If not provided, the default for
								// the provider will be used.
								URL: "",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
