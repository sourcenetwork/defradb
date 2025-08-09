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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithEmptyOrder(t *testing.T) {
	test := testUtils.TestCase{
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

func TestQuerySimpleWithFloat32OrderAscending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					Points: Float32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Points": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Points": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Points": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Points": 19
				}`,
			},
			testUtils.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					HeightM: Float
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 19
				}`,
			},
			testUtils.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					Raw: Blob
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Raw":  "21",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Bob",
					"Raw":  "32",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Carlo",
					"Raw":  "55",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Alice",
					"Raw":  "19",
				},
			},
			testUtils.Request{
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

func TestQuerySimpleWithFloat32OrderDescending(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					Points: Float32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Points": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Points": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Points": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Points": 19
				}`,
			},
			testUtils.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					HeightM: Float
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 55
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 19
				}`,
			},
			testUtils.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `type Users {
					Name: String
					Raw: Blob
				}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Raw":  "21",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Bob",
					"Raw":  "32",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Carlo",
					"Raw":  "55",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Alice",
					"Raw":  "19",
				},
			},
			testUtils.Request{
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

func TestQuerySimple_WithCompoundAliasOrder_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
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
