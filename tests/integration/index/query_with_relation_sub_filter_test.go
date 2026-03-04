// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(0),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Chris"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Discman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Someman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Jumpman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Galaxy",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// 3 indexFetches: process indexes in order "Discman", "Galaxy", "Jumpman", stop when 2 found
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// root: 2 indexFetches (user order), subType: 4 indexFetches (2 Walkman devices per user)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(4).
					WithLevel("root").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Discman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// root: 1 indexFetch (Alice filter), subType: 2 indexFetches (Walkman devices)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(2).
					WithLevel("root").WithIndexFetches(1),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Galaxy",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Pixel",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// 2 indexFetches: 1 for parent filter (users with Walkman) + 1 for Galaxy devices
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPod",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "Pixel",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// 2 indexFetches: 1 for parent filter (Walkman) + 1 for sub-filter (same Walkman)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Toshiba",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Pixel",
					"manufacturer": "Google",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// root (User) has no index
				// subType (Device): 4 indexFetches - 3 for parent filter (3 Walkman devices) + 1 for sub-filter (Sony)
				// Note: For existence checks, we only need 1 match per user, but currently the index fetcher
				// doesn't know about parent relationships. https://github.com/sourcenetwork/defradb/issues/4347
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(0).
					WithLevel("subType").WithIndexFetches(4),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Pixel",
					"manufacturer": "Google",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// root (User) has no index
				// subType (Device): 2 indexFetches for parent filter (2 Walkman devices), sub-filter (manufacturer) applied in-memory
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(0).
					WithLevel("subType").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Aiwa",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPod",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// root (User) has no index
				// subType (Device): 5 indexFetches - 3 device.model fetches (3 Walkman devices)
				// and 2 device.manufacturer fetches (2 Sony devices)
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(0).
					WithLevel("subType").WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
