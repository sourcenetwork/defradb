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

func TestArrayIndex_WithFilterOnIndexedArrayUsingAny_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingAll_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_all: {_geq: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayIndex_WithFilterOnIndexedArrayUsingNone_ShouldNotUseIndex(t *testing.T) {
	req := `query {
		User(filter: {numbers: {_none: {_geq: 33}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// index is not used for _none operator as it might be even less optimal than full scan
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			&action.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 30, 40]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [30, 40, 50, 30]
				}`,
			},
			&action.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [50, 40]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			&action.UpdateDoc{
				DocID: 1,
				Doc: `{
					"name": "Shahzad",
					"numbers": [80, 30, 60]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, 10, 20]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [40, 50]
				}`,
			},
			testUtils.DeleteDoc{DocID: 0},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						booleans: [Boolean!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"booleans": [true, false, true]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"booleans": [false, false]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						booleans: [Boolean] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"booleans": [true, false, true]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"booleans": [false, false]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [4, 3, 7]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						rates: [Float!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"rates": [0.5, 1.0, 1.25]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"rates": [1.5, 1.2]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						rates: [Float] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"rates": [0.5, 1.0, 1.25]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"rates": [1.5, 1.2]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						hobbies: [String] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"hobbies": ["games", "books", "music"]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"hobbies": ["movies", "music"]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [1, 4, 7]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int!] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [3, 4]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [2, 8]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [3, 5, 8]
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.Request{
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
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Islam",
					"numbers": [null, null]
				}`,
			},
			&action.Request{
				Request: `query {
						User(filter: {numbers: {_all: {_geq: 10}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						numbers: [Int] @index
					}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"numbers": [0, null, 2, 3, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"numbers": [10, 20, null]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"numbers": [33, 44, 55]
				}`,
			},
			&action.Request{
				Request: `query {
						User(filter: {numbers: {_none: {_geq: 10}}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
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
