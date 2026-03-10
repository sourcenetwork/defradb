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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestGroupByFieldForTheManySideInCollection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
				{
				  __type(name: "BookField") {
				    name
				    kind
				    enumValues {
				      name
				    }
				  }
				}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"kind": "ENUM",
						"name": "BookField",
						"enumValues": []any{
							// Internal related object fields.
							map[string]any{"name": "author"},
							map[string]any{"name": "_authorID"},

							// Internal fields.
							map[string]any{"name": "_deleted"},
							map[string]any{"name": "GROUP"},
							map[string]any{"name": "_docID"},
							map[string]any{"name": "_version"},

							// User defined collection fields>
							map[string]any{"name": "name"},
							map[string]any{"name": "rating"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGroupByFieldForTheSingleSideInCollection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
				{
				  __type(name: "AuthorField") {
				    name
				    kind
				    enumValues {
				      name
				    }
				  }
				}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"kind": "ENUM",
						"name": "AuthorField",
						"enumValues": []any{
							// Internal related object fields.
							map[string]any{"name": "published"},
							// Note: No `_publishedID` of this side.

							// Internal fields.
							map[string]any{"name": "_deleted"},
							map[string]any{"name": "GROUP"},
							map[string]any{"name": "_docID"},
							map[string]any{"name": "_version"},

							// User defined collection fields>
							map[string]any{"name": "name"},
							map[string]any{"name": "age"},
							map[string]any{"name": "verified"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
