// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfNillableInts(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable ints",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteIntegers: [Int]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, null, 3]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteIntegers": [null, 2, 3, null, 8]
				}`,
			},
			testUtils.Request{
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
