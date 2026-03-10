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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleWithEmbeddings_DoesNotGenerateEmbedding(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
						name_v: [Float32!] @embedding(fields: ["name"], provider: "ollama", model: "nomic-embed-text",  url: "http://localhost:11434/api")
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"Alice"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserView {
							name
							name_v
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "Alice",
							// notice that no embedding has been generated
							"name_v": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
