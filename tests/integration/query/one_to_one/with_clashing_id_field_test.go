// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"
)

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToOneWithClashingIdFieldOnSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation secondary direction, id field with name clash on secondary side",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL will parse the input type as ID and
			// will return an unexpected type error
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author_id: Int
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author_id": 123456
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author_id
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Painted House",
							"author_id": "bae-1a0405fa-e17d-5b0f-8fe2-eb966938df1c",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToOneWithClashingIdFieldOnPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation primary direction, id field with name clash on primary side",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author_id: Int
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
				ExpectedError: "relational id field of invalid kind. Field: author_id, Expected: ID, Actual: Int",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
