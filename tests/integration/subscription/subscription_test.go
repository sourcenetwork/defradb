// Copyright 2023 Democratized Data Foundation
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
	test := testUtils.RequestTestCase{
		Description: "Subscription with user creations",
		Request: `subscription {
					User {
						_key
						name
						age
					}
				}`,
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
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
				Request: `mutation {
					create_User(data: "{\"name\": \"Addo\",\"age\": 31,\"points\": 42.1,\"verified\": true}") {
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
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutation(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Subscription with filter and one user creation",
		Request: `subscription {
					User(filter: {age: {_lt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
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
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutationOutsideFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Subscription with filter and one user creation outside of the filter",
		Request: `subscription {
					User(filter: {age: {_gt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				ExpectedTimout: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithFilterAndCreateMutations(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Subscription with filter and user creation in and outside of the filter",
		Request: `subscription {
					User(filter: {age: {_lt: 30}}) {
						_key
						name
						age
					}
				}`,
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
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
				Request: `mutation {
					create_User(data: "{\"name\": \"Addo\",\"age\": 31,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
				ExpectedTimout: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithUpdateMutations(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Subscription with user creations",
		Request: `subscription {
					User {
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
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					update_User(filter: {name: {_eq: "John"}}, data: "{\"points\": 45}") {
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
	}

	executeTestCase(t, test)
}

func TestSubscriptionWithUpdateAllMutations(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Subscription with user creations",
		Request: `subscription {
					User {
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
		PostSubscriptionRequests: []testUtils.SubscriptionRequest{
			{
				Request: `mutation {
					update_User(data: "{\"points\": 55}") {
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
	}

	executeTestCase(t, test)
}
