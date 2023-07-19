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

func TestQuerySimpleWithIntGreaterThanFilterBlock(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple query with basic filter(age), greater than",
			Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
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
						"Age": 19
					}`,
				},
			},
			Results: []map[string]any{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic filter(age), no results",
			Request: `query {
						Users(filter: {Age: {_gt: 40}}) {
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
			Results: []map[string]any{},
		},
		{
			Description: "Simple query with basic filter(age), multiple results",
			Request: `query {
						Users(filter: {Age: {_gt: 20}}) {
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
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithIntGreaterThanFilterBlockWithNullFilterValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic int greater than filter, with null filter value",
		Request: `query {
					Users(filter: {Age: {_gt: null}}) {
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
		Results: []map[string]any{
			{
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}
