// Copyright 2023 Democratized Data Foundation
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

func TestGroupByFieldForTheManySideInSchema(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test the fields for the many side groupBy are generated.",

		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				  __type(name: "BookFields") {
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
						"name": "BookFields",
						"enumValues": []any{
							// Internal related object fields.
							map[string]any{"name": "author"},
							map[string]any{"name": "author_id"},

							// Internal fields.
							map[string]any{"name": "_deleted"},
							map[string]any{"name": "_group"},
							map[string]any{"name": "_key"},
							map[string]any{"name": "_version"},

							// User defined schema fields>
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

func TestGroupByFieldForTheSingleSideInSchema(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test the fields for the single side groupBy are generated.",

		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				  __type(name: "AuthorFields") {
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
						"name": "AuthorFields",
						"enumValues": []any{
							// Internal related object fields.
							map[string]any{"name": "published"},
							// Note: No `published_id` of this side.

							// Internal fields.
							map[string]any{"name": "_deleted"},
							map[string]any{"name": "_group"},
							map[string]any{"name": "_key"},
							map[string]any{"name": "_version"},

							// User defined schema fields>
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
