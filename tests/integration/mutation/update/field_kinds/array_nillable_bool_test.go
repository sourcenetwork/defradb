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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithArrayOfNillableBooleans(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						likedIndexes: [Boolean]
					}
				`,
			},
			testUtils.CreateDoc{
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
			testUtils.Request{
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
