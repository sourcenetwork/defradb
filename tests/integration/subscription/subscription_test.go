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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSubscriptionWithCreateMutations(t *testing.T) {
	test := testUtils.TestCase{
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
						"User": []map[string]any{
							{
								"_docID": "bae-9591c619-4bca-58eb-8820-28028736ef0c",
								"age":    int64(27),
								"name":   "John",
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"_docID": "bae-45e90427-d499-598b-902a-6a3c65d0b504",
								"age":    int64(31),
								"name":   "Addo",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "Addo",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutation(t *testing.T) {
	test := testUtils.TestCase{
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
						"User": []map[string]any{
							{
								"age":  int64(27),
								"name": "John",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndOneCreateMutationOutsideFilter(t *testing.T) {
	test := testUtils.TestCase{
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
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithFilterAndCreateMutations(t *testing.T) {
	test := testUtils.TestCase{
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
						"User": []map[string]any{
							{
								"age":  int64(27),
								"name": "John",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "Addo",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithUpdateMutations(t *testing.T) {
	test := testUtils.TestCase{
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
						"User": []map[string]any{
							{
								"age":    int64(27),
								"name":   "John",
								"points": float64(45),
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_User(filter: {name: {_eq: "John"}}, input: {points: 45}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscriptionWithUpdateAllMutations(t *testing.T) {
	test := testUtils.TestCase{
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
						"User": []map[string]any{
							{
								"age":    int64(27),
								"name":   "John",
								"points": float64(55),
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"age":    int64(31),
								"name":   "Addo",
								"points": float64(55),
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					update_User(input: {points: 55}) {
						name
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Addo",
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestSubscription_WithDocIDFilter_ShouldOnlyGetUpdatesForThatDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User(docID: "bae-a160ba13-dbf9-50da-a598-018bffa10569") {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"User": []map[string]any{
							{
								"age":  int64(31),
								"name": "Addo",
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"age":  int64(32),
								"name": "Addo",
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
					"age":  27,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Addo",
					"age":  31,
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"age": 28}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc:          `{"age": 32}`,
			},
		},
	}

	execute(t, test)
}

func TestSubscription_WithClose_WontBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User{
						name
						age
					}
				}`,
				Results: nil,
			},
			testUtils.Close{},
		},
	}

	execute(t, test)
}

func TestSubscription_WithCounterCRDT_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						counter: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User {
						counter
					}
				}`,
				Results: []map[string]any{
					{
						"User": []map[string]any{
							{
								"counter": int64(1),
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"counter": int64(2),
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"counter": int64(1),
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"counter": 1}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSubscription_WithDeleteOperation_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.SubscriptionRequest{
				Request: `subscription {
					User (showDeleted: true) { 
						name
						_deleted
					}
				}`,
				Results: []map[string]any{
					{
						"User": []map[string]any{
							{
								"name":     "John",
								"_deleted": false,
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"name":     "Johny",
								"_deleted": false,
							},
						},
					},
					{
						"User": []map[string]any{
							{
								"name":     "Johny",
								"_deleted": true,
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"name": "Johny"}`,
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
