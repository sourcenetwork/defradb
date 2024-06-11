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

func TestQuerySimpleWithIntNotEqualsFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne int filter",
		Request: `query {
					Users(filter: {Age: {_ne: 21}}) {
						Name
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
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne int nil filter",
		Request: `query {
					Users(filter: {Age: {_ne: null}}) {
						Name
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
				`{
					"Name": "Fred"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
			},
			{
				"Name": "Bob",
			},
		},
	}

	executeTestCase(t, test)
}
