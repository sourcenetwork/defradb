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
				Request: clientIntrospectionQuery,
				// TODO: this should pass without error.
				// https://github.com/sourcenetwork/defradb/issues/1502
				ExpectedError: "Unknown kind of type: ",
				// TODO: this should pass without error.
				// https://github.com/sourcenetwork/defradb/issues/1463
				// ExpectedError: "InputFields are missing",
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}
