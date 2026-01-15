// Copyright 2026 Democratized Data Foundation
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

func TestQueryWithIndexOnOneToMany_IfSubFilterOnIndexedField_ShouldFilter(t *testing.T) {
	req := `query {
		User {
			name
			devices(filter: {model: {_eq: "Walkman"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Aiwa"},
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfSubFilterOnNonIndexedField_ShouldNotUseIndex(t *testing.T) {
	req := `query {
		User {
			name
			devices(filter: {manufacturer: {_eq: "Sony"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{"model": "Discman", "manufacturer": "Sony"},
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfSubFilterAndOrderOnIndexedField_ShouldUseIndexForFilter(t *testing.T) {
	req := `query {
		User {
			name
			devices(filter: {model: {_like: "%man"}}, order: {model: ASC}, limit: 2) {
				model
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}
					type Device {
						model: String @index
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Discman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Someman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Jumpman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Galaxy",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{"model": "Discman"},
								{"model": "Jumpman"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 3 indexFetches: process indexes in order "Discman", "Galaxy", "Jumpman", stop when 2 found
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithOrderOnParentAndSubFilter_ShouldFilterPerParent(t *testing.T) {
	req := `query {
		User(order: {name: ASC}) {
			name
			devices(filter: {model: {_eq: "Walkman"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
						{
							"name": "Bob",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Aiwa"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 6 indexFetches: 2 user order + 2 Walkman devices for each user
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(6),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithOrderOnParentAndSubFilter_ShouldFilterBothWithIndexes(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Alice"}}) {
			name
			devices(filter: {model: {_eq: "Walkman"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 3 indexFetches: 1 for parent filter (Alice) + 2 for Walkman devices
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithSameFilterOnParentAndSubType_ShouldFilterBothWithIndexes(t *testing.T) {
	req := `query {
		User(filter: {devices: {model: {_eq: "Walkman"}}}) {
			name
			devices(filter: {model: {_eq: "Galaxy"}}) {
				model
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Galaxy",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Pixel",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Galaxy"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 2 indexFetches: 1 for parent filter (users with Walkman) + 1 for Galaxy devices
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithSameFilterValueOnParentAndSubType_ShouldReturnMatchingDocs(t *testing.T) {
	req := `query {
		User(filter: {devices: {model: {_eq: "Walkman"}}}) {
			name
			devices(filter: {model: {_eq: "Walkman"}}) {
				model
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Pixel",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 2 indexFetches: 1 for parent filter (Walkman) + 1 for sub-filter (same Walkman)
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithParentFilterOnRelationAndSubFilterOnDifferentIndexedField_ShouldUseBothIndexes(t *testing.T) {
	req := `query {
		User(filter: {devices: {model: {_eq: "Walkman"}}}) {
			name
			devices(filter: {manufacturer: {_eq: "Sony"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String @index
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Toshiba",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Pixel",
					"manufacturer": "Google",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 4 indexFetches: 3 for parent filter (3 Walkman devices owned by Alice) + 1 for sub-filter (Sony)
				// Note: For existence checks, we only need 1 match per user, but currently the index fetcher
				// doesn't know about parent relationships. https://github.com/sourcenetwork/defradb/issues/4347
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithParentFilterOnRelationAndSubFilterOnNonIndexedField_ShouldUseParentIndex(t *testing.T) {
	req := `query {
		User(filter: {devices: {model: {_eq: "Walkman"}}}) {
			name
			devices(filter: {manufacturer: {_eq: "Sony"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Pixel",
					"manufacturer": "Google",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 2 indexFetches: for parent filter (2 Walkman devices owned by Alice), sub-filter applied in-memory
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_WithParentFilterOnOwnFieldAndRelationAndSubFilter_ShouldCombineAllFilters(t *testing.T) {
	req := `query {
		User(filter: {name: {_eq: "Alice"}, devices: {model: {_eq: "Walkman"}}}) {
			name
			devices(filter: {manufacturer: {_eq: "Sony"}}) {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}
					type Device {
						model: String @index
						manufacturer: String @index
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"devices": []map[string]any{
								{"model": "Walkman", "manufacturer": "Sony"},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// 5 indexFetches: 3 device.model fetches (3 Walkman devices: 2 Alice, 1 Bob)
				// and 2 device.manufacturer fetches (2 Sony devices)
				// Note: name="Alice" filter is checked after docID lookup (no index)
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
