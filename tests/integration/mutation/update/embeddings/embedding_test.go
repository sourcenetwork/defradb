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

package constraints

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationUpdate_WithMultipleEmbeddingFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Embedding test with updates are currently only compatible with the Go client.
			// The docID is updated by collection.Create after vector embedding generation and
			// the HTTP and CLI clients don't receive that updated docID. This causes the waitForUpdateEvents
			// to fail sinces it receives an update on a docID that wasn't expected. We will look for a solution
			// and update the test accordingly.
			state.GoClientType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			&action.AddDoc{
				// Doc with both embedding fields
				Doc: `{
					"name": "John",
					"about": "He loves fajitas."
				}`,
			},
			&action.UpdateDoc{
				// Doc with both embedding fields
				DocID: 0,
				Doc: `{
					"about": "He loves tacos."
				}`,
			},
			&action.AddDoc{
				// Doc with only one embedding field
				Doc: `{
					"name": "Johnny"
				}`,
			},
			&action.UpdateDoc{
				// Doc with only one embedding field
				DocID: 1,
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `
					query {
						User {
							name_v
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name_v": gomega.And(
								gomega.BeAssignableToTypeOf([]float32{}),
								gomega.HaveLen(768),
							),
						},
						{
							"name_v": gomega.And(
								gomega.BeAssignableToTypeOf([]float32{}),
								gomega.HaveLen(768),
							),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_UserDefinedVectorEmbeddingDoesNotTriggerGeneration_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Embedding test with updates are currently only compatible with the Go client.
			// The docID is updated by collection.Create after vector embedding generation and
			// the HTTP and CLI clients don't receive that updated docID. This causes the waitForUpdateEvents
			// to fail sinces it receives an update on a docID that wasn't expected. We will look for a solution
			// and update the test accordingly.
			state.GoClientType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			&action.AddDoc{
				// Doc with both embedding fields
				Doc: `{
					"name": "John",
					"about": "He loves fajitas."
				}`,
			},
			&action.UpdateDoc{
				DocID: 0,
				Doc: `{
					"name_v": [1, 2, 3]
				}`,
			},
			&action.Request{
				Request: `
					query {
						User {
							name_v
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name_v": []float32{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_FieldsForEmbeddingNotUpdatedDoesNotTriggerGeneration_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// Embedding test with updates are currently only compatible with the Go client.
			// The docID is updated by collection.Create after vector embedding generation and
			// the HTTP and CLI clients don't receive that updated docID. This causes the waitForUpdateEvents
			// to fail sinces it receives an update on a docID that wasn't expected. We will look for a solution
			// and update the test accordingly.
			state.GoClientType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						about: String
						age: Int
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			&action.AddDoc{
				// Doc with both embedding fields
				Doc: `{
					"name": "John",
					"about": "He loves fajitas.",
					"name_v": [1, 2, 3]
				}`,
			},
			&action.UpdateDoc{
				DocID: 0,
				Doc: `{
					"age": 30
				}`,
			},
			&action.Request{
				Request: `
					query {
						User {
							name_v
						}
					}
				`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name_v": []float32{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
