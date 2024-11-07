// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithEmptyOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with empty order",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic order ASC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {Age: ASC}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic order ASC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {CreatedAt: ASC}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic order DESC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {Age: DESC}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic order DESC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {CreatedAt: DESC}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderDescendingAndBooleanOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with compound order",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: [{Age: DESC}, {Verified: ASC}]) {
						Name
						Age
						Verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":     "Carlo",
							"Age":      int64(55),
							"Verified": true,
						},
						{
							"Name":     "Bob",
							"Age":      int64(21),
							"Verified": false,
						},
						{
							"Name":     "John",
							"Age":      int64(21),
							"Verified": true,
						},
						{
							"Name":     "Alice",
							"Age":      int64(19),
							"Verified": false,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithInvalidOrderEnum_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with invalid order",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(order: {Age: INVALID}) {
						Name
						Age
						Verified
					}
				}`,
				ExpectedError: `Argument "order" has invalid value {Age: INVALID}`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMultipleOrderFields_ReturnsError(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Simple query with multiple order fields and a single entry",
			Actions: []any{
				testUtils.Request{
					Request: `query {
					Users(order: {Age: ASC, Name: DESC}) {
						Name
						Age
					}
				}`,
					ExpectedError: "each order argument can only define one field",
				},
			},
		},
		{
			Description: "Simple query with multiple order fields and multiple entries",
			Actions: []any{
				testUtils.Request{
					Request: `query {
					Users(order: [{Age: ASC}, {Age: ASC, Name: DESC}]) {
						Name
						Age
					}
				}`,
					ExpectedError: "each order argument can only define one field",
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimple_WithAliasOrder_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order ASC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {UserAge: ASC}}) {
						Name
						UserAge: Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Alice",
							"UserAge": int64(19),
						},
						{
							"Name":    "John",
							"UserAge": int64(21),
						},
						{
							"Name":    "Bob",
							"UserAge": int64(32),
						},
						{
							"Name":    "Carlo",
							"UserAge": int64(55),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasOrderOnNonAliasedField_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order on non aliased field ASC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {Age: ASC}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasOrderOnNonExistantField_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order on non existant field ASC",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {UserAge: ASC}}) {
						Name
						Age
					}
				}`,
				ExpectedError: `field or alias not found. Name: UserAge`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithInvalidAliasOrder_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order invalid",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {UserAge: invalid}}) {
						Name
						UserAge: Age
					}
				}`,
				ExpectedError: `invalid order direction`,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithEmptyAliasOrder_ShouldDoNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order empty",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: {}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNullAliasOrder_ShouldDoNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order null",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: null}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithIntAliasOrder_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic alias order empty",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(order: {_alias: 1}) {
						Name
						Age
					}
				}`,
				ExpectedError: `invalid order input`,
			},
		},
	}

	executeTestCase(t, test)
}
