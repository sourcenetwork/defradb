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
)

func TestSchemaInlineArrayCreatesSchemaGivenSingleType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "users") {
					name
				}
			}
		`,
		ExpectedData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "users",
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}

func TestSchemaInlineArrayCreatesSchemaGivenSecondType(t *testing.T) {
	test := QueryTestCase{
		Schema: []string{
			`
				type users {
					FavouriteIntegers: [Int!]
				}
			`,
			`
				type books {
					PageNumbers: [Int!]
				}
			`,
		},
		IntrospectionQuery: `
			query IntrospectionQuery {
				__type (name: "books") {
					name
				}
			}
		`,
		ExpectedData: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "books",
			},
		},
	}

	ExecuteQueryTestCase(t, test)
}
