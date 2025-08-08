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

func TestQuerySimpleWithOperationAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with operation alias",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					allUsers: Users {
						_docID
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"allUsers": []map[string]any{
						{
							"_docID": "bae-75cb8b0a-00d7-57c8-8906-29687cbbb15c",
							"Name":   "John",
							"Age":    int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
