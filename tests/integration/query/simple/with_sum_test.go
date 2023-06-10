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

func TestQuerySimpleWithSumOnUndefinedObject(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, sum on undefined object",
		Request: `query {
					_sum
				}`,
		ExpectedError: "aggregate must be provided with a property to aggregate",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSumOnUndefinedField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, sum on undefined field",
		Request: `query {
					_sum(Users: {})
				}`,
		ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSumOnEmptyCollection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, sum on empty",
		Request: `query {
					_sum(Users: {field: Age})
				}`,
		Results: []map[string]any{
			{
				"_sum": int64(0),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSum(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query, sum",
		Request: `query {
					_sum(Users: {field: Age})
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 30
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_sum": int64(51),
			},
		},
	}

	executeTestCase(t, test)
}
