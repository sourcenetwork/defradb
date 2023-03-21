// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaInlineArrayCreatesSchemaGivenSingleType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						FavouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
						__type (name: "users") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "users",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestSchemaInlineArrayCreatesSchemaGivenSecondType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type users {
						FavouriteIntegers: [Int!]
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type books {
						PageNumbers: [Int!]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query IntrospectionQuery {
						__type (name: "books") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "books",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users", "books"}, test)
}
