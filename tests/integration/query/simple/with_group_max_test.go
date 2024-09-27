// Copyright 2024 Democratized Data Foundation
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
	"math"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndMaxOfUndefined_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with max on unspecified field",
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
						_max
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
		Description: "Simple query with group by number, no children, max on non-rendered group, empty collection",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						_max(_group: {field: Age})
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
		Description: "Simple query with group by string, max on non-rendered group integer value",
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
						_max(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_max": int64(38),
						},
						{
							"Name": "Alice",
							"_max": int64(-19),
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
		Description: "Simple query with group by string, max on non-rendered group nil and integer values",
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
						_max(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_max": int64(32),
						},
						{
							"Name": "Alice",
							"_max": int64(19),
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
		Description: "Simple query with group by string, with child group by boolean, and max of max on int",
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
						_max(_group: {field: _max})
						_group (groupBy: [Verified]){
							Verified
							_max(_group: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_max": int64(34),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     int64(32),
								},
								{
									"Verified": false,
									"_max":     int64(34),
								},
							},
						},
						{
							"Name": "Carlo",
							"_max": int64(55),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     int64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"_max": int64(19),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_max":     int64(19),
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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildEmptyFloatMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, max on non-rendered group float (default) value",
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
						_max(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_max": float64(1.89),
						},
						{
							"Name": "Alice",
							"_max": float64(-math.MaxFloat64),
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
		Description: "Simple query with group by string, max on non-rendered group float value",
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
						_max(_group: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_max": float64(1.89),
						},
						{
							"Name": "Alice",
							"_max": float64(2.04),
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
		Description: "Simple query with group by string, with child group by boolean, and max of max on float",
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
						_max(_group: {field: _max})
						_group (groupBy: [Verified]){
							Verified
							_max(_group: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"_max": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_max":     float64(2.04),
								},
							},
						},
						{
							"Name": "John",
							"_max": float64(2.22),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     float64(1.82),
								},
								{
									"Verified": false,
									"_max":     float64(2.22),
								},
							},
						},
						{
							"Name": "Carlo",
							"_max": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     float64(1.74),
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

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfMaxOfFloat_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean, and max of max of max of float",
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
						_max(_group: {field: _max})
						_group (groupBy: [Verified]){
							Verified
							_max(_group: {field: HeightM})
							_group (groupBy: [Age]){
								Age
								_max(_group: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"_max": float64(1.74),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     float64(1.74),
									"_group": []map[string]any{
										{
											"Age":  int64(55),
											"_max": float64(1.74),
										},
									},
								},
							},
						},
						{
							"Name": "Alice",
							"_max": float64(2.04),
							"_group": []map[string]any{
								{
									"Verified": false,
									"_max":     float64(2.04),
									"_group": []map[string]any{
										{
											"Age":  int64(19),
											"_max": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"_max": float64(2.22),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_max":     float64(1.82),
									"_group": []map[string]any{
										{
											"Age":  int64(32),
											"_max": float64(1.61),
										},
										{
											"Age":  int64(25),
											"_max": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"_max":     float64(2.22),
									"_group": []map[string]any{
										{
											"Age":  int64(34),
											"_max": float64(2.22),
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
