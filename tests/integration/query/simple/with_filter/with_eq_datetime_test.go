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

func TestQuerySimpleWithDateTimeEqualsFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter(age)",
		Request: `query {
					Users(filter: {CreatedAt: {_eq: "2017-07-23T03:46:56-05:00"}}) {
						Name
						Age
						CreatedAt
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
					"CreatedAt": "2016-07-23T03:46:56-05:00"
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name":      "John",
					"Age":       int64(21),
					"CreatedAt": testUtils.MustParseTime("2017-07-23T03:46:56-05:00"),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic filter(age)",
		Request: `query {
					Users(filter: {CreatedAt: {_eq: null}}) {
						Name
						Age
						CreatedAt
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
					"CreatedAt": "2016-07-23T03:46:56-05:00"
				}`,
				`{
					"Name": "Fred",
					"Age": 44
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name":      "Fred",
					"Age":       int64(44),
					"CreatedAt": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}
