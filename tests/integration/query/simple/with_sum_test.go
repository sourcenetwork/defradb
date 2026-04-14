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

func TestQuerySimpleWithSumOnUndefinedObject(t *testing.T) {
	test := testUtils.TestCase{
		Description: "SUM with no collection argument returns an error.",
		Actions: []any{
			&action.Request{
				Request: `query {
					SUM
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSumOnUndefinedField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "SUM on a collection without specifying a field returns an error.",
		Actions: []any{
			&action.Request{
				Request: `query {
					SUM(Users: {})
				}`,
				ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSumOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "SUM on an empty collection returns zero.",
		Actions: []any{
			&action.Request{
				Request: `query {
					SUM(Users: {field: Age})
				}`,
				Results: map[string]any{
					"SUM": int64(0),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithSum(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Top-level SUM of an integer field returns the correct total.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
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
					SUM(Users: {field: Age})
				}`,
				Results: map[string]any{
					"SUM": int64(51),
				},
			},
		},
	}

	executeTestCase(t, test)
}
