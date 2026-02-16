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
			&action.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.CreateDoc{
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
