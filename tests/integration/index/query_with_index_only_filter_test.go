// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithIndex_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Islam"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "If there are non-indexed fields in the query, they should be fetched",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Islam",
							"age":  int64(32),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Islam"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfSeveralDocsWithEqFilter_ShouldFetchAll(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Islam"}}) {
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "If there are several docs matching _eq filter, they should be fetched",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Islam",
					"age": 18
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"age": int64(32)},
						{"age": int64(18)},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithGreaterThanFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_gt: 48}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _gt filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithGreaterOrEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_ge: 48}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _ge filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Keenan"},
						{"name": "Chris"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLessThanFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_lt: 22}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _lt filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLessOrEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_le: 23}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _le filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Bruno"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_ne: "Islam"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _ne filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
						{"name": "Bruno"},
						{"name": "Chris"},
						{"name": "Fred"},
						{"name": "John"},
						{"name": "Keenan"},
						{"name": "Roy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithInFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_in: [20, 33]}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithInFilterOnFloat_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						rate: Float @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"rate": 20.0
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"rate": 20.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred",
					"rate": 20.2
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"rate": 20.3
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {rate: {_in: [20, 20.2]}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfSeveralDocsWithInFilter_ShouldFetchAll(t *testing.T) {
	req := `query {
		User(filter: {name: {_in: ["Islam"]}}) {
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "If there are several docs matching _in filter, they should be fetched",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Islam",
					"age": 18
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"age": int64(32)},
						{"age": int64(18)},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotInFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_nin: [20, 23, 28, 33, 42, 55]}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _nin filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
						{"name": "Roy"},
						{"name": "Keenan"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithLikeFilter_ShouldFetch(t *testing.T) {
	req1 := `query {
			User(filter: {email: {_like: "a%"}}) {
				name
			}
		}`
	req2 := `query {
			User(filter: {email: {_like: "%d@gmail.com"}}) {
				name
			}
		}`
	req3 := `query {
			User(filter: {email: {_like: "%e%"}}) {
				name
			}
		}`
	req4 := `query {
		User(filter: {email: {_like: "fred@gmail.com"}}) {
			name
		}
	}`
	req5 := `query {
		User(filter: {email: {_like: "a%@gmail.com"}}) {
			name
		}
	}`
	req6 := `query {
		User(filter: {email: {_like: "a%com%m"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _like filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						email: String @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req3,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Keenan"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req3),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req4,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req4),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req5,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req5),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req6,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req6),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNotLikeFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_nlike: "%h%"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _nlike filter",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
						{"name": "Bruno"},
						{"name": "Fred"},
						{"name": "Islam"},
						{"name": "Keenan"},
						{"name": "Roy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_EmptyFilterOnIndexedField_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"age": 33
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {name: {}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test checks if a query with a filter on 2 relations (one of which is indexed) works.
// Because of 2 relations in the query a parallelNode will be used with each child focusing
// on fetching one of the relations. This test makes sure the result of the second child
// (say Device with manufacturer) doesn't overwrite the result of the first child (say Device with owner).
// Also as the fetching is inverted (because of the index) we fetch first the secondary doc which
// is User and fetch all primary docs (Device) that reference that User. For fetching the primary
// docs we use the same planNode which in this case happens to be multiscanNode (source of parallelNode).
// For every second call multiscanNode will return the result of the first call, but in this case
// we have only one consumer, so take the source of the multiscanNode and use it to fetch the primary docs
// to avoid having all docs doubled.
func TestQueryWithIndex_WithFilterOn2Relations_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}

					type Manufacturer {
						name: String
						devices: [Device]
					}
					
					type Device  {
						owner: User 
						manufacturer: Manufacturer 
						model: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Apple",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"model":           "iPhone",
					"owner_id":        testUtils.NewDocIndex(0, 0),
					"manufacturer_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"model":           "MacBook",
					"owner_id":        testUtils.NewDocIndex(0, 0),
					"manufacturer_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					Device (filter: {
						manufacturer: {name: {_eq: "Apple"}},
						owner: {name: {_eq: "John"}}
					}) {
						model
					}
				}`,
				Results: map[string]any{
					"Device": []map[string]any{
						{
							"model": "iPhone",
						},
						{
							"model": "MacBook",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterAgainstIntField_ShouldFetchNilValues(t *testing.T) {
	req1 := `query {
		User(filter: {age: {_ne: 48}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {age: {_ne: null}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  48,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"age":  nil,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"age":  42,
				},
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterAgainstFloatField_ShouldFetchNilValues(t *testing.T) {
	req1 := `query {
		User(filter: {rating: {_ne: 4.5}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {rating: {_ne: null}}) {
			name	
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						rating: Float @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "John",
					"rating": 4.5,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Andy",
					"rating": nil,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "Shahzad",
					"rating": 4.2,
				},
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterAgainstStringField_ShouldFetchNilValues(t *testing.T) {
	req1 := `query {
		User(filter: {city: {_ne: "Istanbul"}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {city: {_ne: null}}) {
			name	
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						city: String @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"city": "Istanbul",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"city": nil,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"city": "Lucerne",
				},
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterAgainstDateTimeField_ShouldFetchNilValues(t *testing.T) {
	req1 := `query {
		User(filter: {birthdate: {_ne: "2020-01-01T00:00:00Z"}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {birthdate: {_ne: null}}) {
			name	
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						birthdate: DateTime @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":      "John",
					"birthdate": "2020-01-01T00:00:00Z",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":      "Andy",
					"birthdate": nil,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":      "Shahzad",
					"birthdate": "2024-01-01T00:00:00Z",
				},
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithNeFilterAgainstBooleanField_ShouldFetchNilValues(t *testing.T) {
	req1 := `query {
		User(filter: {verified: {_ne: true}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {verified: {_ne: null}}) {
			name	
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						verified: Boolean @index
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":     "John",
					"verified": true,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":     "Andy",
					"verified": nil,
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":     "Shahzad",
					"verified": false,
				},
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
