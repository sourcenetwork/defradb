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

func TestQuerySimpleWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query requesting __typename returns the collection type name for each document.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						__typename
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "John",
							"__typename": "Users",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithAliasedTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query with an aliased __typename returns the type name under the alias.",
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						__typename
						t1: __typename
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":       "John",
							"__typename": "Users",
							"t1":         "Users",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
