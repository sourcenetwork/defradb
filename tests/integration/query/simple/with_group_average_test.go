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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndAverageOfUndefined(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with average on unspecified field",
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
						_avg
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
		Description: "Simple query with group by number, no children, average on non-rendered group, empty collection",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_avg(_group: {field: Age})
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
		Description: "Simple query with group by string, average on non-rendered group integer value",
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
						_avg(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(35),
						},
						{
							"Name": "Alice",
							"_avg": float64(-19),
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
		Description: "Simple query with group by string, average on non-rendered group nil and integer values",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			testUtils.CreateDoc{
				// Age is undefined here and must be ignored
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
						_avg(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(32),
						},
						{
							"Name": "Alice",
							"_avg": float64(19),
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
		Description: "Simple query with group by string, with child group by boolean, and average of average on int",
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
						_avg(_group: {field: _avg})
						_group (groupBy: [Verified]){
							Verified
							_avg(_group: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(31.25),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(28.5),
								},
								{
									"Verified": false,
									"_avg":     float64(34),
								},
							},
						},
						{
							"Name": "Carlo",
							"_avg": float64(55),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"_avg": float64(19),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_avg":     float64(19),
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

func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatAverage(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, average on non-rendered group float (default) value",
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
						_avg(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(1.855),
						},
						{
							"Name": "Alice",
							"_avg": float64(0),
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
		Description: "Simple query with group by string, average on non-rendered group float value",
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
						_avg(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(1.855),
						},
						{
							"Name": "Alice",
							"_avg": float64(2.04),
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
		Description: "Simple query with group by string, with child group by boolean, and average of average on float",
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
						_avg(_group: {field: _avg})
						_group (groupBy: [Verified]){
							Verified
							_avg(_group: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"_avg": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_avg":     float64(2.04),
								},
							},
						},
						{
							"Name": "John",
							"_avg": float64(1.9675000000000002),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(1.715),
								},
								{
									"Verified": false,
									"_avg":     float64(2.22),
								},
							},
						},
						{
							"Name": "Carlo",
							"_avg": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(1.74),
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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfAverageOfFloat(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean, and average of average of average of float",
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
						_avg(_group: {field: _avg})
						_group (groupBy: [Verified]){
							Verified
							_avg(_group: {field: HeightM})
							_group (groupBy: [Age]){
								Age
								_avg(_group: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_avg": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(1.74),
									"_group": []map[string]any{
										{
											"Age":  int64(55),
											"_avg": float64(1.74),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"_avg": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_avg":     float64(2.04),
									"_group": []map[string]any{
										{
											"Age":  int64(19),
											"_avg": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"_avg": float64(1.9675000000000002),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(1.715),
									"_group": []map[string]any{
										{
											"Age":  int64(32),
											"_avg": float64(1.61),
										},
										{
											"Age":  int64(25),
											"_avg": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"_avg":     float64(2.22),
									"_group": []map[string]any{
										{
											"Age":  int64(34),
											"_avg": float64(2.22),
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
