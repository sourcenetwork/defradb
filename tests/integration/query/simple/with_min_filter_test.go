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

func TestQuerySimple_WithMinFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with min filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					_min(Users: {field: Age, filter: {Age: {_gt: 21}}})
				}`,
				Results: map[string]any{
					"_min": int64(30),
				},
			},
		},
	}

	executeTestCase(t, test)
}
