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
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndMinOfUndefined_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users (groupBy: [Name]) {
						Name
						MIN
					}
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMinOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						Age
						MIN(GROUP: {field: Age})
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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 38
				}`,
			},
			&action.CreateDoc{
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
						MIN(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  int64(32),
						},
						{
							"Name": "Alice",
							"MIN":  int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildNilMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
				}`,
			},
			&action.CreateDoc{
				// Age is undefined here
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MIN(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  int64(32),
						},
						{
							"Name": "Alice",
							"MIN":  int64(19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfInt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
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
						MIN(GROUP: {field: MIN})
						GROUP (groupBy: [Verified]){
							Verified
							MIN(GROUP: {field: Age})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  int64(25),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      int64(25),
								},
								{
									"Verified": false,
									"MIN":      int64(34),
								},
							},
						},
						{
							"Name": "Alice",
							"MIN":  int64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MIN":      int64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"MIN":  int64(55),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      int64(55),
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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildEmptyFloatMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MIN(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  float64(1.82),
						},
						{
							"Name": "Alice",
							"MIN":  nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildFloatMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.89
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Alice",
					"HeightM": 2.04
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Name]) {
						Name
						MIN(GROUP: {field: HeightM})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  float64(1.82),
						},
						{
							"Name": "Alice",
							"MIN":  float64(2.04),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfFloat_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
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
						MIN(GROUP: {field: MIN})
						GROUP (groupBy: [Verified]){
							Verified
							MIN(GROUP: {field: HeightM})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"MIN":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      float64(1.74),
								},
							},
						},
						{
							"Name": "John",
							"MIN":  float64(1.61),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      float64(1.61),
								},
								{
									"Verified": false,
									"MIN":      float64(2.22),
								},
							},
						},
						{
							"Name": "Alice",
							"MIN":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MIN":      float64(2.04),
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

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfMinOfFloat_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.82,
					"Age": 25,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 1.61,
					"Age": 32,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.22,
					"Age": 34,
					"Verified": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"Name": "Carlo",
					"HeightM": 1.74,
					"Age": 55,
					"Verified": true
				}`,
			},
			&action.CreateDoc{
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
						MIN(GROUP: {field: MIN})
						GROUP (groupBy: [Verified]){
							Verified
							MIN(GROUP: {field: HeightM})
							GROUP (groupBy: [Age]){
								Age
								MIN(GROUP: {field: HeightM})
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Alice",
							"MIN":  float64(2.04),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"MIN":      float64(2.04),
									"GROUP": []map[string]any{
										{
											"Age": int64(19),
											"MIN": float64(2.04),
										},
									},
								},
							},
						},
						{
							"Name": "John",
							"MIN":  float64(1.61),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      float64(1.61),
									"GROUP": []map[string]any{
										{
											"Age": int64(32),
											"MIN": float64(1.61),
										},
										{
											"Age": int64(25),
											"MIN": float64(1.82),
										},
									},
								},
								{
									"Verified": false,
									"MIN":      float64(2.22),
									"GROUP": []map[string]any{
										{
											"Age": int64(34),
											"MIN": float64(2.22),
										},
									},
								},
							},
						},
						{
							"Name": "Carlo",
							"MIN":  float64(1.74),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"MIN":      float64(1.74),
									"GROUP": []map[string]any{
										{
											"Age": int64(55),
											"MIN": float64(1.74),
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
