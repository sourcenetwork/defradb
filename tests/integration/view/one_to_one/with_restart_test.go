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

func TestView_OneToOneEmbeddedSchemaIsNotLostORestart(t *testing.T) {
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
			// After creating the view, restart and ensure that `BookView` is not forgotten.
			testUtils.Restart{},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "AuthorView") {
							name
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "AuthorView",
					},
				},
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
