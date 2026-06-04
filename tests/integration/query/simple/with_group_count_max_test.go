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

func TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfCount_Succeeds(t *testing.T) {
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
						MAX(GROUP: {field: COUNT})
						GROUP (groupBy: [Verified]){
							Verified
							COUNT(GROUP: {})
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MAX":  int64(2),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"COUNT":    int(2),
								},
								{
									"Verified": false,
									"COUNT":    int(1),
								},
							},
						},
						{
							"Name": "Alice",
							"MAX":  int64(1),
							"GROUP": []map[string]any{
								{
									"Verified": false,
									"COUNT":    int(1),
								},
							},
						},
						{
							"Name": "Carlo",
							"MAX":  int64(1),
							"GROUP": []map[string]any{
								{
									"Verified": true,
									"COUNT":    int(1),
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
