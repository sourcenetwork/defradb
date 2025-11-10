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
	"github.com/sourcenetwork/defradb/tests/lenses"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

func TestView_SimpleWithTransformAggregate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						age: Int
					}
				`,
			},
			testUtils.CreateView{
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
				Transform: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.StandardDeviationModulePath,
							Arguments: map[string]any{
								"src": "age",
								"dst": "stddev",
							},
						},
					},
				}),
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age": 30,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age": 26,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age": 34,
				},
			},
			testUtils.Request{
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
