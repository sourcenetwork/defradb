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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleWithSizeConstraint_CacheLessView_DoesNotErrorOnSizeViolation(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.CachelessViewType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						pointsListInt: [Int!]
						pointsListFloat32: [Float32!]
						pointsListFloat64: [Float64!]
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
						pointsListInt
						pointsListFloat32
						pointsListFloat64
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
						pointsListInt: [Int!] @constraints(size: 2)
						pointsListFloat32: [Float32!] @constraints(size: 2)
						pointsListFloat64: [Float64!] @constraints(size: 2)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Alice",
					"pointsListInt": [1, 2, 3],
					"pointsListFloat32": [1, 2, 3],
					"pointsListFloat64": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							name
							pointsListInt
							pointsListFloat32
							pointsListFloat64
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "Alice",
							// notice the size constraint is not enforced on views
							"pointsListInt":     []int64{1, 2, 3},
							"pointsListFloat32": []float32{1, 2, 3},
							"pointsListFloat64": []float64{1, 2, 3},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents a potential bug with the materialized view where the return type for arrays is
// an interface with float64 values instead of the expected int64, float32 or float64 array types
// TODO: https://github.com/sourcenetwork/defradb/issues/3428
func TestView_SimpleWithSizeConstraint_MaterializedView_DoesNotErrorOnSizeViolation(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						pointsListInt: [Int!]
						pointsListFloat32: [Float32!]
						pointsListFloat64: [Float64!]
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
						pointsListInt
						pointsListFloat32
						pointsListFloat64
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
						pointsListInt: [Int!] @constraints(size: 2)
						pointsListFloat32: [Float32!] @constraints(size: 2)
						pointsListFloat64: [Float64!] @constraints(size: 2)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Alice",
					"pointsListInt": [1, 2, 3],
					"pointsListFloat32": [1, 2, 3],
					"pointsListFloat64": [1, 2, 3]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							name
							pointsListInt
							pointsListFloat32
							pointsListFloat64
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "Alice",
							// notice the size constraint is not enforced on views
							"pointsListInt":     []any{float64(1), float64(2), float64(3)},
							"pointsListFloat32": []any{float64(1), float64(2), float64(3)},
							"pointsListFloat64": []any{float64(1), float64(2), float64(3)},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
