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

func TestQuerySimpleWithIntLessThanFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic lt int filter with greater value",
		Request: `query {
					Users(filter: {Age: {_lt: 22}}) {
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
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntLessThanFilterBlockWithNullValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic lt int filter with null value",
		Request: `query {
					Users(filter: {Age: {_lt: null}}) {
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
					"Name": "Bob"
				}`,
			},
		},
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}
