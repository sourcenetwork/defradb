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

func TestMutationUpdate_WithArrayOfStringsInts(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						preferredStrings: [String]
					}
				`,
			},
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						Users {
							preferredStrings
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
