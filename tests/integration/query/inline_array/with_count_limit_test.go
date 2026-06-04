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

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArrayWithCountWithLimitGreaterThanLength(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(favouriteIntegers: {limit: 3})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Shahzad",
							"COUNT": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithCountWithLimit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"favouriteIntegers": [-1, 2, -1, 1, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						COUNT(favouriteIntegers: {limit: 3})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "Shahzad",
							"COUNT": 3,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
