// Copyright 2025 Democratized Data Foundation
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

	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithMultipleEmbeddingFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with multiple embedding fields",
		SupportedClientTypes: immutable.Some([]testUtils.ClientType{
			// Embedding test with mutations are currently only compatible with the Go client.
			// The docID is updated by collection. Create after vector embedding generation and
			// the HTTP and CLI clients don't receive that updated docID. This causes the waitForUpdateEvents
			// to fail since it receives an update on a docID that wasn't expected. We will look for a solution
			// and update the test accordingly.
			testUtils.GoClientType,
		}),
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
					"about": "He loves tacos.",
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
