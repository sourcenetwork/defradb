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

func TestSimpleMutationUpdateWithBooleanFilterWhereResultFilteredOut(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple update mutation with boolean equals filter",
		// The update will result in a record that no longer matches the filter
		Request: `mutation {
					update_User(filter: {verified: {_eq: true}}, data: "{\"verified\":false}") {
						_key
						name
						points
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
			},
		},
		// As the record no longer matches the filter it is not returned
		Results: []map[string]any{},
	}

	ExecuteTestCase(t, test)
}

func TestSimpleMutationUpdateWithBooleanFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Simple update mutation with boolean equals filter",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
							_key
							name
							points
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
					"name":   "John",
					"points": float64(59),
				},
			},
		},
		{
			Description: "Simple update mutation with boolean equals filter, multiple rows but single match",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
							_key
							name
							points
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1
					}`,
					`{
						"name": "Bob",
						"age": 39,
						"verified": false,
						"points": 66.6
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
					"name":   "John",
					"points": float64(59),
				},
			},
		},
		{
			Description: "Simple update mutation with boolean equals filter, multiple rows",
			Request: `mutation {
						update_User(filter: {verified: {_eq: true}}, data: "{\"points\": 59}") {
							_key
							name
							points
						}
					}`,
			Docs: map[int][]string{
				0: {
					`{
						"name": "John",
						"age": 27,
						"verified": true,
						"points": 42.1
					}`,
					`{
						"name": "Bob",
						"age": 39,
						"verified": true,
						"points": 66.6
					}`,
				},
			},
			Results: []map[string]any{
				{
					"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
					"name":   "John",
					"points": float64(59),
				},
				{
					"_key":   "bae-455b5896-6203-582f-b46e-729c53a2d14b",
					"name":   "Bob",
					"points": float64(59),
				},
			},
		},
	}

	for _, test := range tests {
		ExecuteTestCase(t, test)
	}
}

func TestSimpleMutationUpdateWithIdInFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple update mutation with id in filter, multiple rows",
		Request: `mutation {
					update_User(ids: ["bae-0a24cf29-b2c2-5861-9d00-abd6250c475d", "bae-958c9334-73cf-5695-bf06-cf06826babfa"], data: "{\"points\": 59}") {
						_key
						name
						points
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
				`{
					"name": "Bob",
					"age": 39,
					"verified": false,
					"points": 66.6
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
				"name":   "John",
				"points": float64(59),
			},
			{
				"_key":   "bae-958c9334-73cf-5695-bf06-cf06826babfa",
				"name":   "Bob",
				"points": float64(59),
			},
		},
	}

	ExecuteTestCase(t, test)
}

func TestSimpleMutationUpdateWithIdEqualsFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple update mutation with id equals filter, multiple rows but single match",
		Request: `mutation {
					update_User(id: "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d", data: "{\"points\": 59}") {
						_key
						name
						points
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
				`{
					"name": "Bob",
					"age": 39,
					"verified": false,
					"points": 66.6
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
				"name":   "John",
				"points": float64(59),
			},
		},
	}

	ExecuteTestCase(t, test)
}

func TestSimpleMutationUpdateWithNonExistantId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple update mutation with non existant id",
		Request: `mutation {
					update_User(id: "bae-does-not-exist", data: "{\"points\": 59}") {
						_key
						name
						points
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
			},
		},
		Results: []map[string]any{},
	}

	ExecuteTestCase(t, test)
}
