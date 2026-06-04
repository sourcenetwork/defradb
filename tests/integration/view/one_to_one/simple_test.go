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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_OneToOneDuplicateEmbeddedSchema_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						books: [Book]
					}
					type Book {
						name: String
						author: Author
					}
				`,
			},
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			// Try and create a second view that creates a new `BookView`, this
			// should error as `BookView` has already been created by the first view.
			&action.AddView{
				Query: `
					Author {
						authorName: name
						books {
							bookName: name
						}
					}
				`,
				SDL: `
					type AuthorAliasView @materialized(if: false) {
						authorName: String
						books: [BookView]
					}
					interface BookView {
						bookName: String
					}
				`,
				ExpectedError: "collection already exists. Name: BookView",
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "BookView") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "BookView",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
