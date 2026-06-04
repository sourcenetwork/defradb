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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_SimpleWithTransformAggregate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						age: Int
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.StandardDeviationModulePath,
							Arguments: map[string]any{
								"src": "age",
								"dst": "stddev",
							},
						},
					},
				},
			},
			&action.AddView{
				Query: `
					User {
						age
					}
				`,
				SDL: `
					type UserStdDev @materialized(if: false) {
						stddev: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": 30,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": 26,
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"age": 34,
				},
			},
			&action.Request{
				Request: `
					query {
						UserStdDev {
							stddev
						}
					}
				`,
				Results: map[string]any{
					"UserStdDev": []map[string]any{
						{
							"stddev": float64(3.265986323710904),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
