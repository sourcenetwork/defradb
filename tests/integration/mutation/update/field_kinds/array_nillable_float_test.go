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

func TestMutationUpdate_WithArrayOfNillableFloats(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable floats",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						favouriteFloats: [Float]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, null, -0.00000000001, 10]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"favouriteFloats": [3.1425, -0.00000000001, null, 10]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: []map[string]any{
					{
						"favouriteFloats": []immutable.Option[float64]{
							immutable.Some(3.1425),
							immutable.Some(-0.00000000001),
							immutable.None[float64](),
							immutable.Some[float64](10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
