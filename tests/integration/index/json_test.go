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

func TestJSONIndex_WithFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_eq: 168}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 168}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "John"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithGtFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_gt: 178}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 178}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithGeFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_geq: 178}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 178}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Jesse",
					"custom": null
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Chris",
					"custom": 180
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithLtFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_lt: 178}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 178}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
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
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithLeFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_leq: 178}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 178}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterOnNumberField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_neq: 178}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 178}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": "168 cm"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
						{"name": "John"},
						{"name": "Andy"},
						{"name": "Keenan"},
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

func TestJSONIndex_WithEqFilterOnStringField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {title: {_eq: "Mr"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": "Mr", "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": "Mr"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": 7}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithLikeFilterOnStringField_ShouldUseIndex(t *testing.T) {
	likeReq := `query {
		User(filter: {custom: {title: {_like: "D%"}}}) {
			name
		}
	}`
	ilikeReq := `query {
		User(filter: {custom: {title: {_ilike: "D%"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": "Mr", "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": "dr"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": 7}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			&action.Request{
				Request: likeReq,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(likeReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
			&action.Request{
				Request: ilikeReq,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(ilikeReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNLikeFilterOnStringField_ShouldUseIndex(t *testing.T) {
	nlikeReq := `query {
		User(filter: {custom: {title: {_nlike: "D%"}}}) {
			name
		}
	}`
	nilikeReq := `query {
		User(filter: {custom: {title: {_nilike: "D%"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": "Mr", "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": "dr"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": 7}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			&action.Request{
				Request: nlikeReq,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(nlikeReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
			&action.Request{
				Request: nilikeReq,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(nilikeReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithEqFilterOnBoolField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {isStudent: {_eq: true}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"isStudent": true, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"isStudent": true}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"isStudent": "very much true"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"isStudent": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"isStudent": false}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Islam"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterOnBoolField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {isStudent: {_neq: false}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"isStudent": true, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"isStudent": true}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"isStudent": "very much true"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"isStudent": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"isStudent": false}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
						{"name": "Keenan"},
						{"name": "John"},
						{"name": "Islam"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithEqFilterOnNullField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {title: {_eq: null}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": null, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"title": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": "null"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": 0}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"title": "Dr"}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterOnNullField_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {title: {_neq: null}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"title": null, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"weight": 80, "BMI": 25}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"title": "null"}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"title": 0}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
						{"name": "Keenan"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_UponUpdate_ShouldUseNewIndexValues(t *testing.T) {
	req1 := `query {
		User(filter: {custom: {height: {_eq: 172}}}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {custom: {BMI: {_eq: 22}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "BMI": 25}
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 172, "BMI": 22}
				}`,
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithInFilter_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_in: [168, 180]}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "weight": 80}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"height": 172, "weight": 75}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": 190, "weight": 85}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Fred",
					"custom": {"height": 180, "weight": 70}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Fred"},
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithInFilterOfDifferentTypes_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_in: [168, 180, "172 cm"]}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "weight": 80}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"height": 172, "weight": 75}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": 190, "weight": 85}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Fred",
					"custom": {"height": "172 cm", "weight": 70}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNinFilter_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {custom: {height: {_nin: [168, 180]}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "weight": 80}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"height": 172, "weight": 75}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": 190, "weight": 85}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Keenan"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNotAndInFilter_ShouldNotUseIndex(t *testing.T) {
	req := `query {
		User(filter: {_not: {custom: {height: {_in: [168, 180]}}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "weight": 80}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"height": 172, "weight": 75}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": 190, "weight": 85}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Keenan"},
						{"name": "Shahzad"},
					},
				},
			},
			// we don't assert index usage here because the query is not using the index
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithCompoundFilterCondition_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(filter: {_and: [
			{custom: {height: {_eq: 180}}},
			{custom: {weight: {_eq: 80}}}
		]}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": {"height": 168, "weight": 70}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": {"height": 180, "weight": 80}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Shahzad",
					"custom": {"height": 180, "weight": 75}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Keenan",
					"custom": {"height": 190, "weight": 85}
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterAgainstNumberField_ShouldFetchNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {age: {_neq: 48}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"age": 42,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterAgainstStringField_ShouldFetchNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {city: {_neq: "Istanbul"}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"city": "Istanbul",
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"city": nil,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"city": "Lucerne",
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterAgainstBoolField_ShouldFetchNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {verified: {_neq: true}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"verified": true,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"verified": nil,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"verified": false,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Shahzad"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithNeFilterAgainstNullField_ShouldFetchNonNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {age: {_neq: null}}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"custom": map[string]any{
						"age": 48,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Andy",
					"custom": map[string]any{
						"age": nil,
					},
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"name": "Shahzad",
					"custom": map[string]any{
						"age": 42,
					},
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithEqFilterAgainstExplicitNullField_ShouldFetchNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {_eq: null}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": null
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": 100
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithEqFilterAgainstOmittedNullField_ShouldFetchNullValues(t *testing.T) {
	req := `query {
		User(filter: {custom: {_eq: null}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						custom: JSON @index
					}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Kyle"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Islam",
					"custom": 100
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Kyle"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestJSONIndex_WithGreaterThanFilterOnTopLevelJSONField_ShouldUseIndex(t *testing.T) {
	req := `query {
		Users(filter: {custom: {_gt: 20}}) {
			name
			custom
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						custom: JSON @index
					}
				`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "John",
					"custom": 21
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "David",
					"custom": 19
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Bruno",
					"custom": {"height": null}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Andy",
					"custom": {"height": 190}
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name": "Jesse",
					"custom": null
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"custom": int64(21),
						},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
