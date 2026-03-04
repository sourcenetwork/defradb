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

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				},
			},
			&action.AddView{
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
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.ExplainRequest{
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
