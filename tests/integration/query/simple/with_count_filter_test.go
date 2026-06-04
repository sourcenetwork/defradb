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

func TestQuerySimpleWithCountWithFilter(t *testing.T) {
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
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {filter: {Age: {_gt: 26}}})
				}`,
				Results: map[string]any{
					"COUNT": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCountWithDateTimeFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21,
					"CreatedAt": "2017-07-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30,
					"CreatedAt": "2017-09-23T03:46:56-05:00"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 32,
					"CreatedAt": "2017-10-23T03:46:56-05:00"
				}`,
			},
			&action.Request{
				Request: `query {
					COUNT(Users: {filter: {CreatedAt: {_gt: "2017-08-23T03:46:56-05:00"}}})
				}`,
				Results: map[string]any{
					"COUNT": 2,
				},
			},
		},
	}

	executeTestCase(t, test)
}
