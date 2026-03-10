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

package client_introspection

import (
	_ "embed"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestClientIntrospectionWithOneToManyCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
	testUtils.ExecuteTestCase(t, test)
}
