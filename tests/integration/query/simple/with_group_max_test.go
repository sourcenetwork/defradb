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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndMaxOfUndefined_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users (groupBy: [Name]) {
						Name
						MAX
					}
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMaxOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						MAX(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 38
				}`,
			},
			&action.AddDoc{
				// It is important to test negative values here, due to the auto-typing of numbers
				Doc: `{
					"Name": "Alice",
					"Age": -19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  int64(38),
						},
						{
							"Name": "Alice",
							"MAX":  int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildNilMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				// Age is undefined here
				Doc: `{
					"Name": "John"
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
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  int64(32),
						},
						{
							"Name": "Alice",
							"MAX":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfInt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
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
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: MAX})
						GROUP (groupBy: [Verified]){
							Verified
							MAX(GROUP: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  int64(34),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      int64(32),
								},
								{
									"Verified": false,
									"MAX":      int64(34),
								},
							},
						},
						{
							"Name": "Alice",
							"MAX":  int64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MAX":      int64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"MAX":  int64(55),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      int64(55),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildEmptyFloatMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  float64(1.89),
						},
						{
							"Name": "Alice",
							"MAX":  nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildFloatMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  float64(1.89),
						},
						{
							"Name": "Alice",
							"MAX":  float64(2.04),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfFloat_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: MAX})
						GROUP (groupBy: [Verified]){
							Verified
							MAX(GROUP: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"MAX":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      float64(1.74),
								},
							},
						},
						{
							"Name": "John",
							"MAX":  float64(2.22),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      float64(1.82),
								},
								{
									"Verified": false,
									"MAX":      float64(2.22),
								},
							},
						},
						{
							"Name": "Alice",
							"MAX":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MAX":      float64(2.04),
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfMaxOfFloat_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04,
					"Age": 19,
					"Verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MAX(GROUP: {field: MAX})
						GROUP (groupBy: [Verified]){
							Verified
							MAX(GROUP: {field: HeightM})
							GROUP (groupBy: [Age]){
								Age
								MAX(GROUP: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"MAX":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MAX":      float64(2.04),
									"GROUP": []map[string]any{
										{
											"Age": int64(19),
											"MAX": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"MAX":  float64(2.22),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      float64(1.82),
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
											"MAX": float64(1.61),
										},
										{
											"Age": int64(25),
											"MAX": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"MAX":      float64(2.22),
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
											"MAX": float64(2.22),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"MAX":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MAX":      float64(1.74),
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
											"MAX": float64(1.74),
										},
									},
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
