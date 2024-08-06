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

func TestQuerySimpleWithFloatLessThanFilterBlockWithGreaterValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic lt float filter with greater value",
		Request: `query {
					Users(filter: {HeightM: {_lt: 1.820000000001}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLessThanFilterBlockWithGreaterIntValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic lt float filter with greater int value",
		Request: `query {
					Users(filter: {HeightM: {_lt: 2}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{
				{
					"Name": "Bob",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithFloatLessThanFilterBlockWithNullValue(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with basic lt float filter with null value",
		Request: `query {
					Users(filter: {HeightM: {_lt: null}}) {
						Name
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"HeightM": 2.1
				}`,
				`{
					"Name": "Bob",
					"HeightM": 1.82
				}`,
			},
		},
		Results: map[string]any{
			"Users": []map[string]any{},
		},
	}

	executeTestCase(t, test)
}
