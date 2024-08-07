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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndSumOfUndefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with sum on unspecified field",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (groupBy: [Name]) {
						Name
						_sum
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
		Description: "Simple query with group by number, no children, sum on non-rendered group, empty collection",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_sum(_group: {field: Age})
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
		Description: "Simple query with group by string, sum on non-rendered group integer value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 38
				}`,
			},
			testUtils.CreateDoc{
				// It is important to test negative values here, due to the auto-typing of numbers
				Doc: `{
					"Name": "Alice",
					"Age": -19
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_sum": int64(70),
						},
						{
							"Name": "Alice",
							"_sum": int64(-19),
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
		Description: "Simple query with group by string, sum on non-rendered group nil and integer values",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				// Age is undefined here
				Doc: `{
					"Name": "John"
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
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_sum": int64(32),
						},
						{
							"Name": "Alice",
							"_sum": int64(19),
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
		Description: "Simple query with group by string, with child group by boolean, and sum of sum on int",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
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
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: _sum})
						_group (groupBy: [Verified]){
							Verified
							_sum(_group: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_sum": int64(91),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     int64(57),
								},
								{
									"Verified": false,
									"_sum":     int64(34),
								},
							},
						},
						{
							"Name": "Carlo",
							"_sum": int64(55),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     int64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"_sum": int64(19),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_sum":     int64(19),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatSum(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, sum on non-rendered group float (default) value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_sum": float64(3.71),
						},
						{
							"Name": "Alice",
							"_sum": float64(0),
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
		Description: "Simple query with group by string, sum on non-rendered group float value",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_sum": float64(3.71),
						},
						{
							"Name": "Alice",
							"_sum": float64(2.04),
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
		Description: "Simple query with group by string, with child group by boolean, and sum of sum on float",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: _sum})
						_group (groupBy: [Verified]){
							Verified
							_sum(_group: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"_sum": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_sum":     float64(2.04),
								},
							},
						},
						{
							"Name": "John",
							"_sum": float64(5.65),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     float64(3.43),
								},
								{
									"Verified": false,
									"_sum":     float64(2.22),
								},
							},
						},
						{
							"Name": "Carlo",
							"_sum": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     float64(1.74),
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfSumOfFloat(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean, and sum of sum of sum of float",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Age": 25,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Age": 32,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Age": 34,
					"Verified": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Age": 55,
					"Verified": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04,
					"Age": 19,
					"Verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						_sum(_group: {field: _sum})
						_group (groupBy: [Verified]){
							Verified
							_sum(_group: {field: HeightM})
							_group (groupBy: [Age]){
								Age
								_sum(_group: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_sum": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     float64(1.74),
									"_group": []map[string]any{
										{
											"Age":  int64(55),
											"_sum": float64(1.74),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"_sum": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_sum":     float64(2.04),
									"_group": []map[string]any{
										{
											"Age":  int64(19),
											"_sum": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"_sum": float64(5.65),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_sum":     float64(3.43),
									"_group": []map[string]any{
										{
											"Age":  int64(32),
											"_sum": float64(1.61),
										},
										{
											"Age":  int64(25),
											"_sum": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"_sum":     float64(2.22),
									"_group": []map[string]any{
										{
											"Age":  int64(34),
											"_sum": float64(2.22),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
