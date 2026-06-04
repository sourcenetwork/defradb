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

func TestQuerySimpleWithCountOnUndefined(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					COUNT
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCountOnEmptyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					COUNT(Users: {})
				}`,
				Results: map[string]any{
					"COUNT": 0,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCount(t *testing.T) {
	test := testUtils.TestCase{
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
					COUNT(Users: {})
				}`,
				Results: map[string]any{
					"COUNT": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasedCount_OnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					number: COUNT(Users: {})
				}`,
				Results: map[string]any{
					"number": 0,
				},
			},
		},
	}

	executeTestCase(t, test)
}
