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

	"github.com/sourcenetwork/defradb/tests/action"
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
			Actions: append(
				[]any{
					&action.AddSchema{
						Schema: sameFieldNameGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}

func TestQueryOneToManyWithSameFieldName_SingleSide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"_relationship1ID": "bae-5181bbe5-c134-5e97-8928-30c33d3b83ad"
					}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham"
					}`,
			},
			&action.Request{
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
	}

	executeSameFieldNameTestCase(t, test)
}

func TestQueryOneToManyWithSameFieldName_MultiSide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"_relationship1ID": "bae-5181bbe5-c134-5e97-8928-30c33d3b83ad"
					}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham"
					}`,
			},
			&action.Request{
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
	}

	executeSameFieldNameTestCase(t, test)
}
