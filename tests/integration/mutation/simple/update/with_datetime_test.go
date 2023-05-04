// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSimpleDateTimeMutationUpdateWithBooleanFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple DateTime update mutation with boolean equals filter",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"created_at\": \"2021-07-23T03:46:56.647Z\"}") {
							_key
							name
							created_at
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1,
						"created_at": "2011-07-23T03:46:56.647Z"
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":       "bae-e0374cf9-4e46-5494-bb8a-6dea31912d6b",
					"name":       "John",
					"created_at": "2021-07-23T03:46:56.647Z",
				},
			},
		},
		{
			Description: "Simple DateTime update mutation with boolean equals filter, multiple rows but single match",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"created_at\": \"2021-07-23T03:46:56.647Z\"}") {
							_key
							name
							created_at
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1,
						"created_at": "2011-07-23T03:46:56.647Z"
					}`,
					`{
						"name": "Bob",
						"age": 39,
						"verified": false,
						"points": 66.6,
						"created_at": "2041-07-23T03:46:56.647Z"
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":       "bae-e0374cf9-4e46-5494-bb8a-6dea31912d6b",
					"name":       "John",
					"created_at": "2021-07-23T03:46:56.647Z",
				},
			},
		},
		{
			Description: "Simple DateTime update mutation with boolean equals filter, multiple rows",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"created_at\": \"2021-07-23T03:46:56.647Z\"}") {
							_key
							name
							created_at
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1,
						"created_at": "2011-07-23T03:46:56.647Z"
					}`,
					`{
						"name": "Bob",
						"age": 39,
						"verified": true,
						"points": 66.6,
						"created_at": "2001-07-23T03:46:56.647Z"
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":       "bae-b2f6bd19-56bb-5717-8367-a638e3ca52e0",
					"name":       "Bob",
					"created_at": "2021-07-23T03:46:56.647Z",
				},
				{
					"_key":       "bae-e0374cf9-4e46-5494-bb8a-6dea31912d6b",
					"name":       "John",
					"created_at": "2021-07-23T03:46:56.647Z",
				},
			},
		},
	}

	for _, test := range tests {
		ExecuteTestCase(t, test)
	}
}

func TestSimpleDateTimeMutationUpdateWithIdInFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple DateTime update mutation with id in filter, multiple rows",
		Request: `mutation {
					update_User(ids: ["bae-e0374cf9-4e46-5494-bb8a-6dea31912d6b", "bae-b2f6bd19-56bb-5717-8367-a638e3ca52e0"], data: "{\"created_at\": \"2021-07-23T03:46:56.647Z\"}") {
						_key
						name
						created_at
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1,
					"created_at": "2011-07-23T03:46:56.647Z"
				}`,
				`{
					"name": "Bob",
					"age": 39,
					"verified": true,
					"points": 66.6,
					"created_at": "2001-07-23T03:46:56.647Z"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key":       "bae-b2f6bd19-56bb-5717-8367-a638e3ca52e0",
				"name":       "Bob",
				"created_at": "2021-07-23T03:46:56.647Z",
			},
			{
				"_key":       "bae-e0374cf9-4e46-5494-bb8a-6dea31912d6b",
				"name":       "John",
				"created_at": "2021-07-23T03:46:56.647Z",
			},
		},
	}

	ExecuteTestCase(t, test)
}
