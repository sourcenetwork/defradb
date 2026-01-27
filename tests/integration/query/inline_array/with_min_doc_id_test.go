// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test is meant to provide coverage of the planNode.Prefixes
// func by targeting a specific docID in the parent select.
func TestQueryInlineNillableFloatArray_WithDocIDAndMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"pageRatings": [3.1425, 0.00000000001, 10, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Users(docID: "bae-234d84a8-37f9-57ea-9c53-34c247f3b272") {
						name
						_min(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
							"_min": float64(0.00000000001),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
