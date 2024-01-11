// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_OneToOneDuplicateEmbeddedSchema_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one view and duplicate embedded schema",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
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
			testUtils.CreateView{
				Query: `
					Author {
						authorName: name
						books {
							bookName: name
						}
					}
				`,
				SDL: `
					type AuthorAliasView {
						authorName: String
						books: [BookView]
					}
					interface BookView {
						bookName: String
					}
				`,
				ExpectedError: "schema type already exists. Name: BookView",
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
