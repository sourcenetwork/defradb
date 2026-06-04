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

func TestQuerySimpleWithAverageOnUndefinedObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					AVG
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverageOnUndefinedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					AVG(Users: {})
				}`,
				ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverageOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					AVG(Users: {field: Age})
				}`,
				Results: map[string]any{
					"AVG": float64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAverage(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 28
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.Request{
				Request: `query {
					AVG(Users: {field: Age})
				}`,
				Results: map[string]any{
					"AVG": float64(29),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasedAverage_OnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					average: AVG(Users: {field: Age})
				}`,
				Results: map[string]any{
					"average": float64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}
