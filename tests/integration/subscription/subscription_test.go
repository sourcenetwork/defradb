// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscription

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSubscriptionWithCreateMutations(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with user creations",
		Query: `subscription {
					user {
						_key
						name
						age
					}
				}`,
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						"age":  uint64(27),
						"name": "John",
					},
				},
			},
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"Addo\",\"age\": 31,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-18def051-7f0f-5dc9-8a69-2a5e423f6b55",
						"age":  uint64(31),
						"name": "Addo",
					},
				},
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutation(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with filter and one user creation",
		Query: `subscription {
					user(filter: {age: {_lt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						"age":  uint64(27),
						"name": "John",
					},
				},
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutationOutsideFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with filter and one user creation outside of the filter",
		Query: `subscription {
					user(filter: {age: {_gt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				ExpectedTimout: true,
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndCreateMutations(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with filter and user creation in and outside of the filter",
		Query: `subscription {
					user(filter: {age: {_lt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						"age":  uint64(27),
						"name": "John",
					},
				},
			},
			{
				Query: `mutation {
					create_user(data: "{\"name\": \"Addo\",\"age\": 31,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				ExpectedTimout: true,
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithUpdateMutations(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with user creations",
		Query: `subscription {
					user {
						_key
						name
						age
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
					"name": "Addo",
					"age": 35,
					"verified": true,
					"points": 50
				}`,
			},
		},
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					update_user(filter: {name: {_eq: "John"}}, data: "{\"points\": 45}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						"age":    uint64(27),
						"name":   "John",
						"points": float64(45),
					},
				},
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithUpdateAllMutations(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Subscription with user creations",
		Query: `subscription {
					user {
						_key
						name
						age
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
					"name": "Addo",
					"age": 31,
					"verified": true,
					"points": 50
				}`,
			},
		},
		PostSubscriptionQueries: []testUtils.SubscriptionQuery{
			{
				Query: `mutation {
					update_user(data: "{\"points\": 55}") {
						_key
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_key":   "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
						"age":    uint64(27),
						"name":   "John",
						"points": float64(55),
					},
					{
						"_key":   "bae-cf723876-5c6a-5dcf-a877-ab288eb30d57",
						"age":    uint64(31),
						"name":   "Addo",
						"points": float64(55),
					},
				},
			},
		},
		DisableMapStore: true,
	}

	executeTestCase(t, test)
}
