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

func TestQuery_SimpleWithGroupByStringWithInnerGroupBooleanAndMaxOfAverageOfInt_Succeeds(t *testing.T) {
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
