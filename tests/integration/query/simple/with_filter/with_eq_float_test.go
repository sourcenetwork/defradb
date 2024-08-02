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

func TestQuerySimpleWithFloatEqualsFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic float filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {HeightM: {_eq: 2.1}}) {
						Name
						HeightM
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "John",
							"HeightM": float64(2.1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with basic float nil filter",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"HeightM": 2.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(filter: {HeightM: {_eq: null}}) {
						Name
						HeightM
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":    "Fred",
							"HeightM": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
