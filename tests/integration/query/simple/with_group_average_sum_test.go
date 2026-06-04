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

func TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfCountOfInt(t *testing.T) {
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
						SUM(GROUP: {field: AVG})
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
							"SUM":  float64(62.5),
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
							"SUM":  float64(19),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"AVG":      float64(19),
								},
							},
						},
						{
							"Name": "Carlo",
							"SUM":  float64(55),
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

// Note: this test should follow a different code path to `AVG` on it's own
// utilising the existing `SUM` node instead of adding a new one.  This test cannot
// verify that code path is taken, but it does verfiy that the correct result
// is returned to the consumer in case the more efficient code path is taken.
func TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndSum(t *testing.T) {
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
						SUM(GROUP: {field: Age})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"AVG":  float64(35),
							"SUM":  int64(70),
						},
						{
							"Name": "Alice",
							"AVG":  float64(-19),
							"SUM":  int64(-19),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
