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

func TestQuerySimpleWithKeyFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter (key by filter block)",
		Request: `query {
					Users(filter: {_docID: {_eq: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"}}) {
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
				`{
					"Name": "Bob",
					"Age": 32
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(21),
			},
		},
	}

	executeTestCase(t, test)
}
