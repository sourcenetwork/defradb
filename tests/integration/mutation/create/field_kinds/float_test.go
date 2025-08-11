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

func TestMutationCreateFieldKinds_WithFloat(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			testUtils.Request{
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

func TestMutationCreateFieldKinds_WithFloat32(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						points: Float32
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			testUtils.Request{
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

func TestMutationCreateFieldKinds_WithFloat64(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						points: Float64
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"points": 1.2,
				},
			},
			testUtils.Request{
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
