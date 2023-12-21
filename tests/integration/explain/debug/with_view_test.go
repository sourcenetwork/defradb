// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_debug

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var viewPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"selectNode": dataMap{
				"viewNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithView(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (debug) request with view",
		Actions: []any{
			testUtils.SchemaUpdate{
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
					type UserView {
						name: String
					}
				`,
			},
			testUtils.ExplainRequest{
				Request: `query @explain(type: debug) {
					UserView {
						name
					}
				}`,
				ExpectedPatterns: []dataMap{viewPattern},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
