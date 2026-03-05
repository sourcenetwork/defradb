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

package constraints

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAdd_WithSizeConstrain_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddCollection{
				SDL: `
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
