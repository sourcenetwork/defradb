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

func TestQuerySimpleWithDateTimeLEFilterBlockWithEqualValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic le DateTime filter with equal value",
		Request: `query {
					Users(filter: {CreatedAt: {_le: "2017-07-23T03:46:56-05:00"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
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

func TestQuerySimpleWithDateTimeLEFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic le DateTime filter with greater value",
		Request: `query {
					Users(filter: {CreatedAt: {_le: "2018-07-23T03:46:56-05:00"}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
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

func TestQuerySimpleWithDateTimeLEFilterBlockWithNullValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic le DateTime filter with null value",
		Request: `query {
					Users(filter: {CreatedAt: {_le: null}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
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
