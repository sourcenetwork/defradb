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

func TestQuerySimpleWithDateTimeLTFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic lt DateTime filter with equal value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_lt: "2017-07-25T03:46:56-05:00"}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeLTFilterBlockWithNullValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic lt DateTime filter with null value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_lt: null}}) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNilDateTimeLTAndNonNilFilterBlock_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic filter with nil value and non-nil filter",
		Actions: []any{
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name":      "John",
					"Age":       int64(21),
					"CreatedAt": "2017-07-23T03:46:56-05:00",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name":      "Bob",
					"Age":       int64(32),
					"CreatedAt": "2016-07-23T03:46:56-05:00",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Fred",
					"Age":  44,
				},
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {CreatedAt: {_lt: "2017-07-23T03:46:56-05:00"}}) {
						Name
						Age
						CreatedAt
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{

						{
							"Name":      "Bob",
							"Age":       int64(32),
							"CreatedAt": testUtils.MustParseTime("2016-07-23T03:46:56-05:00"),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
