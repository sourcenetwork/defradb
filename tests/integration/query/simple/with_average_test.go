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

func TestQuerySimpleWithAverageOnUndefinedObject(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average on undefined object",
		Request: `query {
					_avg
				}`,
		ExpectedError: "aggregate must be provided with a property to aggregate",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverageOnUndefinedField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average on undefined field",
		Request: `query {
					_avg(Users: {})
				}`,
		ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverageOnEmptyCollection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average on empty",
		Request: `query {
					_avg(Users: {field: Age})
				}`,
		Results: map[string]any{
			"_avg": float64(0),
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverage(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, average",
		Request: `query {
					_avg(Users: {field: Age})
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 28
				}`,
				`{
					"Name": "Bob",
					"Age": 30
				}`,
			},
		},
		Results: map[string]any{
			"_avg": float64(29),
		},
	}

	executeTestCase(t, test)
}
