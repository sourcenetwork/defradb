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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/integration/collection_version"
)

func TestView_OneToMany_GQLIntrospectionTest(t *testing.T) {
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
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "AuthorView") {
							name
							fields {
								name
								type {
									name
									kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "AuthorView",
						"fields": collection_version.DefaultViewObjFields.Append(
							collection_version.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Append(
							collection_version.Field{
								"name": "books",
								"type": map[string]any{
									"kind": "LIST",
									"name": nil,
								},
							},
						).Tidy(),
					},
				},
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "BookView") {
							name
							fields {
								name
								type {
									name
									kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "BookView",
						// Note: `_docID`, `_version`, `_deleted`, etc should not be present,
						// although aggregates and `GROUP` should be.
						// There should also be no `Author` field - the relationship field
						// should only exist on the parent.
						"fields": collection_version.DefaultViewObjFields.Append(
							collection_version.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
