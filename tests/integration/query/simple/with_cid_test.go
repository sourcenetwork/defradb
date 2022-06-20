// Copyright 2022 Democratized Data Foundation
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

// This test is for documentation reasons only. I do not see this
// as desired behaviour.
func TestQuerySimpleWithInvalidCid(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple query with cid",
		Query: `query {
					users (cid: "any non-nil string value - this will be ignored") {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]interface{}{
			{
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}
