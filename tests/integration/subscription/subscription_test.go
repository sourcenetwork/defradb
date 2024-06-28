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
	test := testUtils.TestCase{
		Description: "Subscription with user creations",
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User {
						_docID
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-b3ce089b-f543-5984-be9f-ad7d08969f4e",
						"age":    int64(27),
						"name":   "John",
					},
					{
						"_docID": "bae-bc20b854-10b3-5408-b28c-f273ddda9434",
						"age":    int64(31),
						"name":   "Addo",
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscription with filter and one user creation",
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User(filter: {age: {_lt: 30}}) {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"age":  int64(27),
						"name": "John",
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutationOutsideFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscription with filter and one user creation outside of the filter",
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User(filter: {age: {_gt: 30}}) {
						_docID
						name
						age
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndCreateMutations(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscription with filter and user creation in and outside of the filter",
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User(filter: {age: {_lt: 30}}) {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"age":  int64(27),
						"name": "John",
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithUpdateMutations(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscription with user creations and single mutation",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Addo",
					"age": 35,
					"verified": true,
					"points": 50
				}`,
			},
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User {
						name
						age
						points
					}
				}`,
				Results: []map[string]any{
					{
						"age":    int64(27),
						"name":   "John",
						"points": float64(45),
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_User(filter: {name: {_eq: "John"}}, input: {points: 45}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithUpdateAllMutations(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Subscription with user creations and mutations for all",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27,
					"verified": true,
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Addo",
					"age": 31,
					"verified": true,
					"points": 50
				}`,
			},
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User {
						name
						age
						points
					}
				}`,
				Results: []map[string]any{
					{
						"age":    int64(31),
						"name":   "Addo",
						"points": float64(55),
					},
					{
						"age":    int64(27),
						"name":   "John",
						"points": float64(55),
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_User(input: {points: 55}) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
					{
						"name": "John",
					},
				},
			},
		},
	}

	execute(t, test)
}
