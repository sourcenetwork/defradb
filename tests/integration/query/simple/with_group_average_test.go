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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndAverageOfUndefined(t *testing.T) {
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
						AVG
					}
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						AVG(GROUP: {field: Age})
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverage(t *testing.T) {
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
						AVG(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(35),
						},
						{
							"Name": "Alice",
							"AVG":  float64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildNilAverage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				// Age is undefined here and must be ignored
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
						AVG(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(32),
						},
						{
							"Name": "Alice",
							"AVG":  float64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfInt(t *testing.T) {
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
						AVG(GROUP: {field: AVG})
						GROUP (groupBy: [Verified]){
							Verified
							AVG(GROUP: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(31.25),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(28.5),
								},
								{
									"Verified": false,
									"AVG":      float64(34),
								},
							},
						},
						{
							"Name": "Alice",
							"AVG":  float64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"AVG":      float64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"AVG":  float64(55),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(55),
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatAverage(t *testing.T) {
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
						AVG(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(1.855),
						},
						{
							"Name": "Alice",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildFloatAverage(t *testing.T) {
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
						AVG(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(1.855),
						},
						{
							"Name": "Alice",
							"AVG":  float64(2.04),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfFloat(t *testing.T) {
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
						AVG(GROUP: {field: AVG})
						GROUP (groupBy: [Verified]){
							Verified
							AVG(GROUP: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"AVG":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(1.74),
								},
							},
						},
						{
							"Name": "John",
							"AVG":  float64(1.9675000000000002),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(1.715),
								},
								{
									"Verified": false,
									"AVG":      float64(2.22),
								},
							},
						},
						{
							"Name": "Alice",
							"AVG":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"AVG":      float64(2.04),
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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfAverageOfFloat(t *testing.T) {
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
						AVG(GROUP: {field: AVG})
						GROUP (groupBy: [Verified]){
							Verified
							AVG(GROUP: {field: HeightM})
							GROUP (groupBy: [Age]){
								Age
								AVG(GROUP: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"AVG":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"AVG":      float64(2.04),
									"GROUP": []map[string]any{
										{
											"Age": int64(19),
											"AVG": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"AVG":  float64(1.9675000000000002),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(1.715),
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
											"AVG": float64(1.61),
										},
										{
											"Age": int64(25),
											"AVG": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"AVG":      float64(2.22),
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
											"AVG": float64(2.22),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"AVG":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"AVG":      float64(1.74),
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
											"AVG": float64(1.74),
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
