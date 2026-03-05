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

func TestMutationUpdate_IfFloatFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						rate: Float
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"rate": 0.55
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"rate": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							rate
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"rate": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
