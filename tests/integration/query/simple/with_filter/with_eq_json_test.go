// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithEqOpOnJSONField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "John",
					"custom": "{\"tree\": \"maple\", \"age\": 250}",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Andy",
					"custom": "{\"tree\": \"oak\", \"age\": 450}",
				},
			},
			testUtils.Request{
				// the filtered-by JSON has no spaces, because this is now it's stored.
				Request: `query {
					Users(filter: {custom: {_eq: "{\"tree\":\"oak\",\"age\":450}"}}) {
						name
					}
				}`,
				Results: []map[string]any{{"name": "Andy"}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
