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

func TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMinWithLimit_Succeeds(t *testing.T) {
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
				Doc: `{
					"Name": "John",
					"Age": 28
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
						MIN(GROUP: {field: Age, limit: 2})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"MIN":  int64(28),
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
