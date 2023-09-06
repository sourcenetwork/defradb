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

func TestMutationUpdate_WithArrayOfStringsInts(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, nillable string",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						preferredStrings: [String]
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["", "the previous", "the first", "empty string", null]
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"preferredStrings": ["", "the previous", null, "empty string", "blank string", "hitchi"]
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: []map[string]any{
					{
						"preferredStrings": []immutable.Option[string]{
							immutable.Some(""),
							immutable.Some("the previous"),
							immutable.None[string](),
							immutable.Some("empty string"),
							immutable.Some("blank string"),
							immutable.Some("hitchi"),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
