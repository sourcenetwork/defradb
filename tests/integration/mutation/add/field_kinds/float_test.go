// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddFieldKinds_WithFloat(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float32
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float32(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						points: Float64
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			&action.Request{
				Request: `query {
					User {
						points
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"points": float64(1.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
