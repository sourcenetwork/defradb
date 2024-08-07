// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var sameFieldNameGQLSchema = (`
	type Book {
		name: String
		relationship1: Author
	}

	type Author {
		name: String
		relationship1: [Book]
	}
`)

func executeSameFieldNameTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Description: test.Description,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: sameFieldNameGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}

func TestQueryOneToManyWithSameFieldName(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-many relation query from one side, same field name",
			Actions: []any{
				testUtils.CreateDoc{
					CollectionID: 0,
					Doc: `{
						"name": "Painted House",
						"relationship1_id": "bae-ee5973cf-73c3-558f-8aec-8b590b8e77cf"
					}`,
				},
				testUtils.CreateDoc{
					CollectionID: 1,
					Doc: `{
						"name": "John Grisham"
					}`,
				},
				testUtils.Request{
					Request: `query {
						Book {
							name
							relationship1 {
								name
							}
						}
					}`,
					Results: map[string]any{
						"Book": []map[string]any{
							{
								"name": "Painted House",
								"relationship1": map[string]any{
									"name": "John Grisham",
								},
							},
						},
					},
				},
			},
		},
		{
			Description: "One-to-many relation query from many side, same field name",
			Actions: []any{
				testUtils.CreateDoc{
					CollectionID: 0,
					Doc: `{
						"name": "Painted House",
						"relationship1_id": "bae-ee5973cf-73c3-558f-8aec-8b590b8e77cf"
					}`,
				},
				testUtils.CreateDoc{
					CollectionID: 1,
					Doc: `{
						"name": "John Grisham"
					}`,
				},
				testUtils.Request{
					Request: `query {
						Author {
							name
							relationship1 {
								name
							}
						}
					}`,

					Results: map[string]any{
						"Author": []map[string]any{
							{
								"name": "John Grisham",
								"relationship1": []map[string]any{
									{
										"name": "Painted House",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeSameFieldNameTestCase(t, test)
	}
}
