// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package json

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryJSON_WithAggregateFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					name: String
					custom: JSON
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {
						"tree": "maple",
						"age": 250
					}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {
						"tree": "oak",
						"age": 450
					}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					_count(Users: {filter: {custom: {tree: {_eq: "oak"}}}})
				}`,
				Results: map[string]any{
					"_count": 1,
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
