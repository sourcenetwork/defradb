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
	"math"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimple_WithMaxOnUndefinedObject_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query max on undefined object",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_max
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMaxOnUndefinedField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query max on undefined field",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_max(Users: {})
				}`,
				ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMaxOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query max on empty",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_max(Users: {field: Age})
				}`,
				Results: map[string]any{
					"_max": math.MinInt64,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMax_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query max",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			testUtils.Request{
				Request: `query {
					_max(Users: {field: Age})
				}`,
				Results: map[string]any{
					"_max": int64(30),
				},
			},
		},
	}

	executeTestCase(t, test)
}
