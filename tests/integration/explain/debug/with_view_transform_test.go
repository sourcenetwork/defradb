// Copyright 2024 Democratized Data Foundation
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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

var transformViewPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"selectNode": dataMap{
						"viewNode": dataMap{
							"lensNode": dataMap{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"scanNode": dataMap{},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestDebugExplainRequestWithViewWithTransform(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{testUtils.CachelessViewType}),
		Description:        "Explain (debug) request with view with transform",
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
					type UserView @materialized(if: false) {
						fullName: String
					}
				`,
				Transform: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				}),
			},
			testUtils.ExplainRequest{
				Request: `query @explain(type: debug) {
					UserView {
						fullName
					}
				}`,
				ExpectedPatterns: transformViewPattern,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
