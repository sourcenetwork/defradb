// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client_introspection

import (
	_ "embed"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

//go:embed altair_graphiql_postman_2023.gql
var clientIntrospectionQuery string

func TestClientIntrospectionBasic(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.ClientIntrospectionRequest{
				Request: clientIntrospectionQuery,
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{}, test)
}

// TODO: This should pass without error, but we errors.
func TestClientIntrospectionWithOneToManySchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				type Book {
					name: String
					author: Author
				}
				type Author {
					name: String
					published: [Book]
				}
				`,
			},
			testUtils.ClientIntrospectionRequest{
				Request:       clientIntrospectionQuery,
				ExpectedError: "Unknown kind of type: ",
				// ExpectedError: "InputFields are missing",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Book", "Author"}, test)
}
