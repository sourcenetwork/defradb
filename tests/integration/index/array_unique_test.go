// Copyright 2025 Democratized Data Foundation
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

func TestArrayUniqueIndex_UponDocCreationWithUniqueElement_Succeed(t *testing.T) {
	req := `query {
		User(filter: {nfts: {_any: {_eq: 30}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
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

func TestArrayUniqueIndex_UponDocCreationWithUniqueElements_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Andy",
					"nfts": [50, 30]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueIndex_UponDocUpdateWithUniqueElements_Succeed(t *testing.T) {
	req := `query {
		User(filter: {nfts: {_any: {_eq: 60}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"nfts": [10, 60]
				}`,
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

func TestArrayUniqueIndex_UponDocUpdateWithArrayElementThatExists_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User  {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"nfts": [50, 30]
				}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueIndex_UponDeletingDoc_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User  {
						name: String 
						nfts: [Int!] @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"nfts": [0, 30, 20]
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"nfts": [10, 40]
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestArrayUniqueIndex_WithNilElementsAndAnyOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index(unique: true)
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

func TestArrayUniqueIndex_WithNilElementsAndAllOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index(unique: true)
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

func TestArrayUniqueIndex_WithNilElementsAndNoneOp_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						numbers: [Int] @index(unique: true)
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
