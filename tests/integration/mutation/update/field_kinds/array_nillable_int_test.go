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

func TestMutationUpdate_WithArrayOfNillableInts(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteIntegers: [Int]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, null, 3]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [null, 2, 3, null, 8]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							favouriteIntegers
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"favouriteIntegers": []immutable.Option[int64]{
								immutable.None[int64](),
								immutable.Some[int64](2),
								immutable.Some[int64](3),
								immutable.None[int64](),
								immutable.Some[int64](8),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
