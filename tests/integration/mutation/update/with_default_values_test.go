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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithDefaultValues_DoesNotOverwrite(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String 
						score: Int @default(int: 100)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"score": 0
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						score
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Fred",
							"score": int64(0),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
