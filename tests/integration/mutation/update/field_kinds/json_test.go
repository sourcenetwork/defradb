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

func TestMutationUpdate_IfJSONFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						custom: JSON
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"custom": {"foo": "bar"}
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"custom": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							custom
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
