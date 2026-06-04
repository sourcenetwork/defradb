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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndSumOfUndefined(t *testing.T) {
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
						SUM
					}
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSumOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						SUM(GROUP: {field: Age})
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSum(t *testing.T) {
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
						SUM(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  int64(70),
						},
						{
							"Name": "Alice",
							"SUM":  int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildNilSum(t *testing.T) {
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
						SUM(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  int64(32),
						},
						{
							"Name": "Alice",
							"SUM":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfInt(t *testing.T) {
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
						SUM(GROUP: {field: SUM})
						GROUP (groupBy: [Verified]){
							Verified
							SUM(GROUP: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  int64(91),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      int64(57),
								},
								{
									"Verified": false,
									"SUM":      int64(34),
								},
							},
						},
						{
							"Name": "Alice",
							"SUM":  int64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"SUM":      int64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"SUM":  int64(55),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      int64(55),
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatSum(t *testing.T) {
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
						SUM(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  float64(3.71),
						},
						{
							"Name": "Alice",
							"SUM":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildFloatSum(t *testing.T) {
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
						SUM(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"SUM":  float64(3.71),
						},
						{
							"Name": "Alice",
							"SUM":  float64(2.04),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfFloat(t *testing.T) {
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
						SUM(GROUP: {field: SUM})
						GROUP (groupBy: [Verified]){
							Verified
							SUM(GROUP: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"SUM":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      float64(1.74),
								},
							},
						},
						{
							"Name": "John",
							"SUM":  float64(5.65),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      float64(3.43),
								},
								{
									"Verified": false,
									"SUM":      float64(2.22),
								},
							},
						},
						{
							"Name": "Alice",
							"SUM":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"SUM":      float64(2.04),
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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfSumOfFloat(t *testing.T) {
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
						SUM(GROUP: {field: SUM})
						GROUP (groupBy: [Verified]){
							Verified
							SUM(GROUP: {field: HeightM})
							GROUP (groupBy: [Age]){
								Age
								SUM(GROUP: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"SUM":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"SUM":      float64(2.04),
									"GROUP": []map[string]any{
										{
											"Age": int64(19),
											"SUM": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"SUM":  float64(5.65),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      float64(3.43),
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
											"SUM": float64(1.61),
										},
										{
											"Age": int64(25),
											"SUM": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"SUM":      float64(2.22),
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
											"SUM": float64(2.22),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"SUM":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"SUM":      float64(1.74),
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
											"SUM": float64(1.74),
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
