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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfNillableBooleans(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						likedIndexes: [Boolean]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, true, false, true]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"likedIndexes": [true, true, false, true, null]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							likedIndexes
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"likedIndexes": []immutable.Option[bool]{
								immutable.Some(true),
								immutable.Some(true),
								immutable.Some(false),
								immutable.Some(true),
								immutable.None[bool](),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
