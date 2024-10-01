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

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimple_WithMinOnUndefinedObject_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query min on undefined object",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_min
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinOnUndefinedField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query min on undefined field",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_min(Users: {})
				}`,
				ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query min on empty",
		Actions: []any{
			testUtils.Request{
				Request: `query {
					_min(Users: {field: Age})
				}`,
				Results: map[string]any{
					"_min": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query min",
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
					_min(Users: {field: Age})
				}`,
				Results: map[string]any{
					"_min": int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinAndMaxValueInt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GraphQL does not support 64 bit int
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Description: "Simple query min and max value int",
		Actions: []any{
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Age":  int64(math.MaxInt64),
				},
			},
			testUtils.Request{
				Request: `query {
					_max(Users: {field: Age})
				}`,
				Results: map[string]any{
					"_max": int64(math.MaxInt64),
				},
			},
		},
	}

	executeTestCase(t, test)
}
