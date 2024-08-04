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
	test := testUtils.RequestTestCase{
		Description: "Simple query with operation alias",
		Request: `query {
					allUsers: Users {
						_docID
						Name
						Age
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
		Results: map[string]any{
			"allUsers": []map[string]any{
				{
					"_docID": "bae-d4303725-7db9-53d2-b324-f3ee44020e52",
					"Name":   "John",
					"Age":    int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}
