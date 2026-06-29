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

func TestMutationUpdate_WithArrayOfNillableFloats(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						favouriteFloats: [Float]
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [3.1425, null, -0.00000000001, 10]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"favouriteFloats": [3.1425, -0.00000000001, null, 10]
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							favouriteFloats
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
