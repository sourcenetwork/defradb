// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package constraints

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAdd_WithSizeConstrain_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						numbers: [Int!] @constraints(size: 2)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [27, 28]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							numbers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":    "John",
							"numbers": []int64{27, 28},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithSizeConstrainMismatch_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						numbers: [Int!] @constraints(size: 2)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [27, 28, 29]
				}`,
				ExpectedError: "array size mismatch",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
