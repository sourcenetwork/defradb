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

func TestMutationUpdate_IfIntFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"age": 33
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"age": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
