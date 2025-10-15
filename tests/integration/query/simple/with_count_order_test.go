// Copyright 2025 Democratized Data Foundation
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

// It doesn't matter if we order using ASC or DESC, because the count will always
// be the same. But this test shows that the syntax is valid.
func TestQuerySimpleWithCountWithOrder_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Age": 30,
					"HeightM": 1.8
				}`,
			}, // Count: 2

			testUtils.CreateDoc{
				Doc: `{
					"Age": 25,
					"HeightM": 1.6
				}`,
			}, // Count: 2
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {total: DESC}}) {
						total: _count(HeightM: {}, Age: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"total": 2,
						},
						{
							"total": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
