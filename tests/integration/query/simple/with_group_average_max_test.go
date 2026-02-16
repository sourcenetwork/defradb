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

func TestQuery_SimpleWithGroupByStringWithInnerGroupBooleanAndMaxOfAverageOfInt_Succeeds(t *testing.T) {
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
						MAX(GROUP: {field: AVG})
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
							"MAX":  float64(34),
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
							"MAX":  float64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"AVG":      float64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"MAX":  float64(55),
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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndMax_Succeeds(t *testing.T) {
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
						AVG(GROUP: {field: Age})
						MAX(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(35),
							"MAX":  int64(38),
						},
						{
							"Name": "Alice",
							"AVG":  float64(-19),
							"MAX":  int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
