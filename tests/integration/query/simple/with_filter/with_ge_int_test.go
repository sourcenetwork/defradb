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

func TestQuerySimpleWithIntGEFilterBlockWithEqualValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge int filter with equal value",
		Request: `query {
					Users(filter: {Age: {_ge: 32}}) {
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
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntGEFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge int filter with greater value",
		Request: `query {
					Users(filter: {Age: {_ge: 31}}) {
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
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithIntGEFilterBlockWithNilValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic ge nil filter",
		Request: `query {
					Users(filter: {Age: {_ge: null}}) {
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
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Bob",
				},
				{
					"Name": "John",
				},
			},
		},
	}

	executeTestCase(t, test)
}
