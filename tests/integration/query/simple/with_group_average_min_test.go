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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuery_SimpleWithGroupByStringWithInnerGroupBooleanAndMinOfAverageOfInt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, with child group by boolean, and min of average on int",
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
						_min(_group: {field: _avg})
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
							"_min": float64(28.5),
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
							"_min": float64(55),
							"_group": []map[string]any{
								{
									"Verified": true,
									"_avg":     float64(55),
								},
							},
						},
						{
							"Name": "Alice",
							"_min": float64(19),
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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with group by string, average and min on non-rendered group integer value",
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
						_min(_group: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_avg": float64(35),
							"_min": int64(32),
						},
						{
							"Name": "Alice",
							"_avg": float64(-19),
							"_min": int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
