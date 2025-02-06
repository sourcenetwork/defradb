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

func TestMutationUpdate_WithMultipleEmbeddingFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation",
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
					"about": "He loves fajitas."
				}`,
			},
			testUtils.UpdateDoc{
				// Doc with both embedding fields
				DocID: 0,
				Doc: `{
					"about": "He loves tacos."
				}`,
			},
			testUtils.CreateDoc{
				// Doc with only one embedding field
				Doc: `{
					"name": "Johnny"
				}`,
			},
			testUtils.UpdateDoc{
				// Doc with only one embedding field
				DocID: 1,
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

func TestMutationUpdate_UserDefinedVectorEmbeddingDoesNotTriggerGeneration_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation with manually defined vector embedding",
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
					"about": "He loves fajitas."
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"name_v": [1, 2, 3]
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
		Description: "Simple update mutation with manually defined vector embedding",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						about: String
						age: Int
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			testUtils.CreateDoc{
				// Doc with both embedding fields
				Doc: `{
					"name": "John",
					"about": "He loves fajitas.",
					"name_v": [1, 2, 3]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"age": 30
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
							"name_v": []float32{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
