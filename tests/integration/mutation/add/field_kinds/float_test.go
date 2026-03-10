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
