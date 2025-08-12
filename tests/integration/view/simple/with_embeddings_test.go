// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleWithEmbeddings_DoesNotGenerateEmbedding(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
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
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Alice"
				}`,
			},
			testUtils.Request{
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
