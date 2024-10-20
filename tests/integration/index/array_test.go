// Copyright 2024 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestArrayIndex_WithFilterOnIndexedArrayUsingAny_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingAll_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_all: {_ge: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(9),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingNone_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_none: {_ge: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(9),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateRearrangesArrayElements_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 30, 40]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateRemovesSoughtElement_ShouldNotFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 40]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndexUpdate_IfUpdateAddsSoughtElement_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [80, 30, 60]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
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

func TestArrayIndexDelete_IfUpdateRemovesSoughtElement_ShouldNotFetch(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_gt: 0}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			testUtils.DeleteDoc{DocID: 0},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_Bool_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {booleans: {_any: {_eq: true}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						booleans: [Boolean!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"booleans": [true, false, true]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"booleans": [false, false]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_OptionalBool_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {booleans: {_any: {_eq: true}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						booleans: [Boolean] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"booleans": [true, false, true]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"booleans": [false, false]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_OptionalInt_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 3}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [4, 3, 7]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayIndex_Float_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {rates: {_any: {_eq: 1.25}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						rates: [Float!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"rates": [0.5, 1.0, 1.25]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"rates": [1.5, 1.2]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayIndex_OptionalFloat_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {rates: {_any: {_eq: 1.25}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						rates: [Float] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"rates": [0.5, 1.0, 1.25]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"rates": [1.5, 1.2]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayIndex_OptionalString_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {hobbies: {_any: {_eq: "books"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						hobbies: [String] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"hobbies": ["games", "books", "music"]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"hobbies": ["movies", "music"]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayIndex_WithAnyAndInOperator_Succeed(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_in: [3, 4, 5]}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [1, 4, 7]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
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

func TestArrayIndex_WithAllAndInOperator_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_all: {_in: [3, 4, 5]}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithNoneAndInOperator_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_none: {_in: [4, 5]}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithNoneAndNinOperator_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_none: {_nin: [3, 4, 5]}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithAllAndNinOperator_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_all: {_nin: [4, 5]}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithAnyAndNinOperator_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			testUtils.Request{
				Request: `query {
					User(filter: {numbers: {_any: {_nin: [3, 4, 5]}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithNilElementsAndAnyOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_any: {_eq: 2}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_any: {_eq: null}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithNilElementsAndAllOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam",
					"numbers": [null, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_all: {_ge: 10}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_all: {_eq: null}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithNilElementsAndNoneOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_none: {_ge: 10}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						User(filter: {numbers: {_none: {_eq: null}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
