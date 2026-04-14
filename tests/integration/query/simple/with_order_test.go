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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithEmptyOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Passing an empty order object returns all documents in default order.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithNumericOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a numeric field ascending returns them from lowest to highest.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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

func TestQuerySimpleWithFloat32OrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a float32 field ascending returns them from lowest to highest.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Points: Float32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Points": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Points": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Points": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Points": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {Points: ASC}) {
						Name
						Points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "Alice",
							"Points": float32(19),
						},
						{
							"Name":   "John",
							"Points": float32(21),
						},
						{
							"Name":   "Bob",
							"Points": float32(32),
						},
						{
							"Name":   "Carlo",
							"Points": float32(55),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithFloat64OrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a float64 field ascending returns them from lowest to highest.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					HeightM: Float
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {HeightM: ASC}) {
						Name
						HeightM
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Alice",
							"HeightM": float64(19),
						},
						{
							"Name":    "John",
							"HeightM": float64(21),
						},
						{
							"Name":    "Bob",
							"HeightM": float64(32),
						},
						{
							"Name":    "Carlo",
							"HeightM": float64(55),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithBlobOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a blob field ascending returns them in lexicographic order.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Raw: Blob
				}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Raw":  "21",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Bob",
					"Raw":  "32",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Carlo",
					"Raw":  "55",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Alice",
					"Raw":  "19",
				},
			},
			&action.Request{
				Request: `query {
					Users(order: {Raw: ASC}) {
						Name
						Raw
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Raw":  "19",
						},
						{
							"Name": "John",
							"Raw":  "21",
						},
						{
							"Name": "Bob",
							"Raw":  "32",
						},
						{
							"Name": "Carlo",
							"Raw":  "55",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a DateTime field ascending returns them chronologically.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
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
		Description: "Order documents by a numeric field descending returns them from highest to lowest.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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

func TestQuerySimpleWithFloat32OrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a float32 field descending returns them from highest to lowest.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Points: Float32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Points": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Points": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Points": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Points": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {Points: DESC}) {
						Name
						Points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "Carlo",
							"Points": float32(55),
						},
						{
							"Name":   "Bob",
							"Points": float32(32),
						},
						{
							"Name":   "John",
							"Points": float32(21),
						},
						{
							"Name":   "Alice",
							"Points": float32(19),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWitFloat64OrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a float64 field descending returns them from highest to lowest.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					HeightM: Float
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {HeightM: DESC}) {
						Name
						HeightM
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Carlo",
							"HeightM": float64(55),
						},
						{
							"Name":    "Bob",
							"HeightM": float64(32),
						},
						{
							"Name":    "John",
							"HeightM": float64(21),
						},
						{
							"Name":    "Alice",
							"HeightM": float64(19),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithBlobOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a blob field descending returns them in reverse lexicographic order.",
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users {
					Name: String
					Raw: Blob
				}`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Raw":  "21",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Bob",
					"Raw":  "32",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Carlo",
					"Raw":  "55",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Alice",
					"Raw":  "19",
				},
			},
			&action.Request{
				Request: `query {
					Users(order: {Raw: DESC}) {
						Name
						Raw
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Raw":  "55",
						},
						{
							"Name": "Bob",
							"Raw":  "32",
						},
						{
							"Name": "John",
							"Raw":  "21",
						},
						{
							"Name": "Alice",
							"Raw":  "19",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithDateTimeOrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Order documents by a DateTime field descending returns them reverse-chronologically.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2021-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32,
					"CreatedAt": "2032-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"CreatedAt": "2055-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"CreatedAt": "2019-07-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
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
		Description: "Order by numeric descending then boolean ascending returns the correct compound sort.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
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

func TestQuerySimple_WithMultipleOrderFieldsASCAndASC_ShouldOrderCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Two ASC order fields produce the expected compound sort.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 22
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 24
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: [{Name: ASC}, {Age: ASC}]) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(22),
						},
						{
							"Name": "Alice",
							"Age":  int64(24),
						},
						{
							"Name": "Alice",
							"Age":  int64(38),
						},
						{
							"Name": "Bob",
							"Age":  int64(30),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMultipleOrderFieldsACSAndDESC_ShouldOrderCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Description: "First field ASC, second field DESC produces the expected compound sort.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 22
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 24
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: [{Name: ASC}, {Age: DESC}]) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"Age":  int64(38),
						},
						{
							"Name": "Alice",
							"Age":  int64(24),
						},
						{
							"Name": "Alice",
							"Age":  int64(22),
						},
						{
							"Name": "Bob",
							"Age":  int64(30),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMultipleOrderFieldsDESCAndASC_ShouldOrderCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Description: "First field DESC, second field ASC produces the expected compound sort.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 22
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 24
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: [{Name: DESC}, {Age: ASC}]) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(30),
						},
						{
							"Name": "Alice",
							"Age":  int64(22),
						},
						{
							"Name": "Alice",
							"Age":  int64(24),
						},
						{
							"Name": "Alice",
							"Age":  int64(38),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMultipleOrderFieldsDECSAndDESC_ShouldOrderCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Two DESC order fields produce the expected compound sort.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 22
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 24
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: [{Name: DESC}, {Age: DESC}]) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(30),
						},
						{
							"Name": "Alice",
							"Age":  int64(38),
						},
						{
							"Name": "Alice",
							"Age":  int64(24),
						},
						{
							"Name": "Alice",
							"Age":  int64(22),
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
		Description: "Passing an invalid order direction enum value returns a schema validation error.",
		Actions: []any{
			&action.Request{
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
	test := testUtils.TestCase{
		Description: "Specifying multiple order fields in a single order object returns an error.",
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(order: {Age: ASC, Name: DESC}) {
						Name
						Age
					}
				}`,
				ExpectedError: "each order argument can only define one field",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMultipleOrderFieldsNestedWithinMultpleFields_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Nesting multiple order fields within a compound order object returns an error.",
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(order: [{Age: ASC}, {Age: ASC, Name: DESC}]) {
						Name
						Age
					}
				}`,
				ExpectedError: "each order argument can only define one field",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasOrder_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Ordering by an aliased aggregate field sorts documents by the aggregated value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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
		Description: "Ordering by an alias that refers to a non-aliased field sorts by the underlying field value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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
		Description: "Ordering by an alias that references a non-existent field returns a schema error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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
		Description: "Ordering by an alias with an invalid direction value returns a schema validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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
		Description: "Ordering by an alias with an empty order object returns documents in default order.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {_alias: {}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithNullAliasOrder_ShouldDoNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Ordering by an alias with a null direction returns documents in default order.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {_alias: null}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithIntAliasOrder_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Ordering by an alias with an integer direction value returns a schema validation error.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
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

func TestQuerySimple_WithCompoundAliasOrder_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Ordering by a compound alias that aggregates multiple fields sorts by the combined value.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 21,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: [{_alias: {userAge: DESC}}, {_alias: {isVerified: ASC}}]) {
						Name
						userAge: Age
						isVerified: Verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "Carlo",
							"userAge":    int64(55),
							"isVerified": true,
						},
						{
							"Name":       "Bob",
							"userAge":    int64(21),
							"isVerified": false,
						},
						{
							"Name":       "John",
							"userAge":    int64(21),
							"isVerified": true,
						},
						{
							"Name":       "Alice",
							"userAge":    int64(19),
							"isVerified": false,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
